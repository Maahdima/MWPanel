package service

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"mikrotik-wg-go/dataservice/model"
	"time"
)

var (
	// TODO : read from config environment variables
	accessSecret  = []byte("access_secret")
	refreshSecret = []byte("refresh_secret")

	accessTokenTTL  = time.Minute * 15
	refreshTokenTTL = time.Hour * 24 * 7
)

type Authentication struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewAuthentication(db *gorm.DB) *Authentication {
	return &Authentication{
		db:     db,
		logger: zap.L().Named("AuthenticationService"),
	}
}

func (a *Authentication) Login(username, password string) (accessToken, refreshToken string, expiresIn int64, err error) {
	var admin model.Admin

	if err := a.db.First(&admin, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			a.logger.Error("user not found", zap.String("username", username))
			return "", "", 0, errors.New("user not found")
		}
		a.logger.Error("failed to query user from database", zap.Error(err))
		return "", "", 0, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		a.logger.Error("password mismatch", zap.String("username", username), zap.Error(err))
		return "", "", 0, errors.New("password mismatch")
	}

	accessToken, err = a.generateAccessToken(username)
	if err != nil {
		a.logger.Error("failed to generate access token", zap.Error(err))
		return "", "", 0, err
	}

	refreshToken, err = a.generateRefreshToken(username)
	if err != nil {
		a.logger.Error("failed to generate refresh token", zap.Error(err))
		return "", "", 0, err
	}

	expiresIn = int64(accessTokenTTL.Seconds())
	return accessToken, refreshToken, expiresIn, nil
}

func (a *Authentication) generateAccessToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(accessTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(accessSecret)
}

func (a *Authentication) generateRefreshToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(refreshTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(refreshSecret)
}
