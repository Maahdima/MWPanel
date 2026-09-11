package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/utils"
)

const (
	totpIssuer          = "MWPanel"
	totpPendingTokenTTL = 5 * time.Minute
	totpPendingTokenTyp = "2fa_pending"
)

var (
	ErrTotpAlreadyEnabled = errors.New("two-factor authentication is already enabled")
	ErrTotpNotEnabled     = errors.New("two-factor authentication is not enabled")
	ErrTotpNotPending     = errors.New("two-factor setup has not been started")
	ErrInvalidTotpCode    = errors.New("invalid authentication code")
	ErrInvalidTempToken   = errors.New("invalid or expired verification token")
)

func (a *Authentication) Login(username, password string) (*schema.LoginResponse, error) {
	admin, err := a.authenticatePassword(username, password)
	if err != nil {
		return nil, err
	}

	if admin.TotpEnabled {
		tempToken, err := a.generate2FAPendingToken(admin.Username)
		if err != nil {
			a.logger.Error("failed to generate 2FA pending token", zap.Error(err))
			return nil, err
		}
		return &schema.LoginResponse{
			Requires2FA: true,
			TempToken:   tempToken,
		}, nil
	}

	return a.issueLoginTokens(admin)
}

func (a *Authentication) Verify2FA(tempToken, code string) (*schema.LoginResponse, error) {
	username, err := a.parse2FAPendingToken(tempToken)
	if err != nil {
		return nil, err
	}

	var admin model.Admin
	if err := a.db.First(&admin, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		a.logger.Error("failed to query user from database", zap.Error(err))
		return nil, err
	}

	if !admin.TotpEnabled || admin.TotpSecret == nil {
		return nil, ErrTotpNotEnabled
	}

	ok, err := a.verifyTotpOrRecovery(&admin, code)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrInvalidTotpCode
	}

	return a.issueLoginTokens(&admin)
}

func (a *Authentication) GetTotpStatus(username string) (*schema.TotpStatusResponse, error) {
	admin, err := a.findAdminByUsername(username)
	if err != nil {
		return nil, err
	}
	return &schema.TotpStatusResponse{Enabled: admin.TotpEnabled}, nil
}

func (a *Authentication) StartTotpSetup(username string) (*schema.TotpSetupResponse, error) {
	admin, err := a.findAdminByUsername(username)
	if err != nil {
		return nil, err
	}
	if admin.TotpEnabled {
		return nil, ErrTotpAlreadyEnabled
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      totpIssuer,
		AccountName: admin.Username,
	})
	if err != nil {
		a.logger.Error("failed to generate TOTP key", zap.Error(err))
		return nil, err
	}

	secret := key.Secret()
	admin.TotpPendingSecret = &secret
	if err := a.db.Save(admin).Error; err != nil {
		a.logger.Error("failed to save pending TOTP secret", zap.Error(err))
		return nil, err
	}

	img, err := key.Image(200, 200)
	if err != nil {
		a.logger.Error("failed to generate TOTP QR image", zap.Error(err))
		return nil, err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		a.logger.Error("failed to encode TOTP QR image", zap.Error(err))
		return nil, err
	}

	return &schema.TotpSetupResponse{
		Secret:        secret,
		OTPAuthURL:    key.URL(),
		QRCodeDataURL: "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}

func (a *Authentication) ConfirmTotpSetup(username, code string) (*schema.TotpConfirmResponse, error) {
	admin, err := a.findAdminByUsername(username)
	if err != nil {
		return nil, err
	}
	if admin.TotpEnabled {
		return nil, ErrTotpAlreadyEnabled
	}
	if admin.TotpPendingSecret == nil || *admin.TotpPendingSecret == "" {
		return nil, ErrTotpNotPending
	}

	if !totp.Validate(strings.TrimSpace(code), *admin.TotpPendingSecret) {
		return nil, ErrInvalidTotpCode
	}

	plainCodes, hashedJSON, err := utils.GenerateRecoveryCodes()
	if err != nil {
		a.logger.Error("failed to generate recovery codes", zap.Error(err))
		return nil, err
	}

	admin.TotpSecret = admin.TotpPendingSecret
	admin.TotpPendingSecret = nil
	admin.TotpEnabled = true
	admin.TotpRecoveryCodes = &hashedJSON

	if err := a.db.Save(admin).Error; err != nil {
		a.logger.Error("failed to enable TOTP", zap.Error(err))
		return nil, err
	}

	return &schema.TotpConfirmResponse{RecoveryCodes: plainCodes}, nil
}

func (a *Authentication) DisableTotp(username, password, code string) error {
	admin, err := a.authenticatePassword(username, password)
	if err != nil {
		return err
	}
	if !admin.TotpEnabled || admin.TotpSecret == nil {
		return ErrTotpNotEnabled
	}

	ok, err := a.verifyTotpOrRecovery(admin, code)
	if err != nil {
		return err
	}
	if !ok {
		return ErrInvalidTotpCode
	}

	admin.TotpEnabled = false
	admin.TotpSecret = nil
	admin.TotpPendingSecret = nil
	admin.TotpRecoveryCodes = nil

	if err := a.db.Save(admin).Error; err != nil {
		a.logger.Error("failed to disable TOTP", zap.Error(err))
		return err
	}
	return nil
}

func (a *Authentication) authenticatePassword(username, password string) (*model.Admin, error) {
	var admin model.Admin
	if err := a.db.First(&admin, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			a.logger.Error("user not found", zap.String("username", username))
			return nil, gorm.ErrRecordNotFound
		}
		a.logger.Error("failed to query user from database", zap.Error(err))
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		a.logger.Error("password mismatch", zap.String("username", username), zap.Error(err))
		return nil, errors.New("password mismatch")
	}
	return &admin, nil
}

func (a *Authentication) findAdminByUsername(username string) (*model.Admin, error) {
	var admin model.Admin
	if err := a.db.First(&admin, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			a.logger.Error("user not found", zap.String("username", username))
			return nil, gorm.ErrRecordNotFound
		}
		a.logger.Error("failed to query user from database", zap.Error(err))
		return nil, err
	}
	return &admin, nil
}

func (a *Authentication) issueLoginTokens(admin *model.Admin) (*schema.LoginResponse, error) {
	accessToken, err := a.generateAccessToken(admin.Username)
	if err != nil {
		a.logger.Error("failed to generate access token", zap.Error(err))
		return nil, err
	}

	refreshToken, err := a.generateRefreshToken(admin.Username)
	if err != nil {
		a.logger.Error("failed to generate refresh token", zap.Error(err))
		return nil, err
	}

	return &schema.LoginResponse{
		Requires2FA:  false,
		UserID:       admin.ID,
		Username:     admin.Username,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTokenTTL.Seconds()),
	}, nil
}

func (a *Authentication) generate2FAPendingToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"sub": username,
		"typ": totpPendingTokenTyp,
		"exp": time.Now().Add(totpPendingTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.RefreshSecret)
}

func (a *Authentication) parse2FAPendingToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.RefreshSecret, nil
	})
	if err != nil || !token.Valid {
		return "", ErrInvalidTempToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidTempToken
	}

	typ, _ := claims["typ"].(string)
	if typ != totpPendingTokenTyp {
		return "", ErrInvalidTempToken
	}

	username, _ := claims["sub"].(string)
	if username == "" {
		return "", ErrInvalidTempToken
	}
	return username, nil
}

func (a *Authentication) verifyTotpOrRecovery(admin *model.Admin, code string) (bool, error) {
	normalized := strings.TrimSpace(code)
	if admin.TotpSecret != nil && totp.Validate(normalized, *admin.TotpSecret) {
		return true, nil
	}

	used, err := a.consumeRecoveryCode(admin, normalized)
	if err != nil {
		return false, err
	}
	return used, nil
}

func (a *Authentication) consumeRecoveryCode(admin *model.Admin, code string) (bool, error) {
	if admin.TotpRecoveryCodes == nil || *admin.TotpRecoveryCodes == "" {
		return false, nil
	}

	var hashes []string
	if err := json.Unmarshal([]byte(*admin.TotpRecoveryCodes), &hashes); err != nil {
		a.logger.Error("failed to parse recovery codes", zap.Error(err))
		return false, err
	}

	normalized := utils.NormalizeRecoveryCode(code)
	remaining := make([]string, 0, len(hashes))
	matched := false

	for _, hash := range hashes {
		if matched {
			remaining = append(remaining, hash)
			continue
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(normalized)) == nil {
			matched = true
			continue
		}
		remaining = append(remaining, hash)
	}

	if !matched {
		return false, nil
	}

	encoded, err := json.Marshal(remaining)
	if err != nil {
		return false, err
	}
	encodedStr := string(encoded)
	admin.TotpRecoveryCodes = &encodedStr
	if err := a.db.Save(admin).Error; err != nil {
		a.logger.Error("failed to update recovery codes", zap.Error(err))
		return false, err
	}
	return true, nil
}
