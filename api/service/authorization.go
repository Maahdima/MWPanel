package service

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/maahdima/mwp/api/dataservice/model"
	"gorm.io/gorm"
)

var (
	ErrUnauthorizedAccount = errors.New("unauthorized account")
	ErrRoleForbidden       = errors.New("this operation requires superadmin access")
)

type Principal struct {
	AccountID  uint   `json:"account_id"`
	Username   string `json:"username"`
	Role       string `json:"role"`
	ResellerID *uint  `json:"reseller_id,omitempty"`
}

// ResolvePrincipal reads current state on every request. JWT claims or payloads
// cannot preserve revoked roles or choose another reseller's ownership.
func (a *Authentication) ResolvePrincipal(username string) (Principal, error) {
	var account model.Admin
	if err := a.db.Where("username = ? AND is_active = ?", username, true).First(&account).Error; err != nil {
		return Principal{}, ErrUnauthorizedAccount
	}
	switch account.Role {
	case model.RoleSuperadmin:
		if account.ResellerID != nil {
			return Principal{}, ErrUnauthorizedAccount
		}
	case model.RoleReseller:
		if account.ResellerID == nil || *account.ResellerID == 0 {
			return Principal{}, ErrUnauthorizedAccount
		}
		var tenant model.Reseller
		if err := a.db.Where("is_active = ?", true).First(&tenant, *account.ResellerID).Error; err != nil {
			return Principal{}, ErrUnauthorizedAccount
		}
	default:
		return Principal{}, ErrUnauthorizedAccount
	}
	return Principal{account.ID, account.Username, account.Role, account.ResellerID}, nil
}

func (a *Authentication) ParseAccessToken(raw string) (*jwt.Token, Principal, error) {
	token, err := jwt.Parse(raw, func(token *jwt.Token) (any, error) { return a.AccessSecret, nil }, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil || token == nil || !token.Valid {
		return nil, Principal{}, ErrUnauthorizedAccount
	}
	username, err := token.Claims.GetSubject()
	if err != nil || username == "" {
		return nil, Principal{}, ErrUnauthorizedAccount
	}
	principal, err := a.ResolvePrincipal(username)
	return token, principal, err
}

func (p Principal) PeerScope(db *gorm.DB) *gorm.DB {
	if p.Role == model.RoleSuperadmin && p.AccountID != 0 && p.ResellerID == nil {
		return db
	}
	if p.Role == model.RoleReseller && p.AccountID != 0 && p.ResellerID != nil && *p.ResellerID != 0 {
		return db.Where("reseller_id = ?", *p.ResellerID)
	}
	return db.Where("1 = 0")
}

func (p Principal) FindOwnedPeer(db *gorm.DB, id uint) (model.Peer, error) {
	var peer model.Peer
	err := p.PeerScope(db).First(&peer, "id = ?", id).Error
	return peer, err
}
