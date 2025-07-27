package service

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/maahdima/mwp/api/config"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/utils"
)

var (
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
)

func init() {
	authCfg := config.GetAuthConfig()

	accessTTL, _ := strconv.Atoi(authCfg.AccessTokenTTL)
	refreshTTL, _ := strconv.Atoi(authCfg.RefreshTokenTTL)

	accessTokenTTL = time.Duration(accessTTL) * time.Second
	refreshTokenTTL = time.Duration(refreshTTL) * time.Second
}

type Authentication struct {
	db            *gorm.DB
	AccessSecret  []byte
	RefreshSecret []byte
	logger        *zap.Logger
}

func NewAuthentication(db *gorm.DB) *Authentication {
	return &Authentication{
		db:            db,
		AccessSecret:  []byte(utils.RandomString(24)),
		RefreshSecret: []byte(utils.RandomString(24)),
		logger:        zap.L().Named("AuthenticationService"),
	}
}

func (a *Authentication) Login(username, password string) (*schema.LoginResponse, error) {
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

	accessToken, err := a.generateAccessToken(username)
	if err != nil {
		a.logger.Error("failed to generate access token", zap.Error(err))
		return nil, err
	}

	refreshToken, err := a.generateRefreshToken(username)
	if err != nil {
		a.logger.Error("failed to generate refresh token", zap.Error(err))
		return nil, err
	}

	expiresIn := int64(accessTokenTTL.Seconds())
	return &schema.LoginResponse{
		UserID:       admin.ID,
		Username:     admin.Username,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

func (a *Authentication) UpdateProfile(oldUsername, oldPassword string, newUsername, newPassword *string) error {
	var admin model.Admin

	if err := a.db.First(&admin, "username = ?", oldUsername).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			a.logger.Error("user not found", zap.String("username", oldUsername))
			return gorm.ErrRecordNotFound
		}
		a.logger.Error("failed to query user from database", zap.Error(err))
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(oldPassword)); err != nil {
		a.logger.Error("password mismatch", zap.String("username", oldUsername), zap.Error(err))
		return errors.New("password mismatch")
	}

	if newUsername != nil {
		admin.Username = *newUsername
	}

	if newPassword != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*newPassword), bcrypt.DefaultCost)
		if err != nil {
			a.logger.Error("failed to hash new password", zap.Error(err))
			return err
		}
		admin.Password = string(hashedPassword)
	}

	if err := a.db.Save(&admin).Error; err != nil {
		a.logger.Error("failed to update user profile", zap.Error(err))
		return err
	}

	return nil
}

func (a *Authentication) generateAccessToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(accessTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.AccessSecret)
}

func (a *Authentication) generateRefreshToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(refreshTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.RefreshSecret)
}
