package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/maahdima/mwp/api/dataservice/model"
	"gorm.io/gorm"
)

type PaymentInstructions struct {
	RateRial     string `json:"rate_rial"`
	CardNumber   string `json:"card_number"`
	Cardholder   string `json:"cardholder"`
	Instructions string `json:"instructions"`
}

func (b *Billing) Instructions(actor Principal) (PaymentInstructions, error) {
	if !billingAdmin(actor) && !(actor.AccountID != 0 && actor.Role == model.RoleReseller && actor.ResellerID != nil) {
		return PaymentInstructions{}, ErrRoleForbidden
	}

	var settings model.BillingSettings

	err := b.db.First(&settings, 1).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return PaymentInstructions{RateRial: strconv.FormatInt(DefaultPricePerGiB30DaysRial, 10)}, nil
	}

	if err != nil {
		return PaymentInstructions{}, err
	}

	return PaymentInstructions{strconv.FormatInt(settings.RateRial, 10), settings.CardNumber, settings.Cardholder, settings.Instructions}, nil
}

func (b *Billing) UpdateInstructions(actor Principal, input PaymentInstructions) error {
	if !billingAdmin(actor) {
		return ErrRoleForbidden
	}

	rate, err := strconv.ParseInt(input.RateRial, 10, 64)

	if err != nil || rate < 0 {
		return ErrInvalidBillingQuantity
	}

	card := strings.TrimSpace(input.CardNumber)

	if card != "" {
		if len(card) != 16 || strings.TrimSpace(input.Cardholder) == "" {
			return ErrInvalidBillingQuantity
		}

		for _, digit := range card {
			if digit < '0' || digit > '9' {
				return ErrInvalidBillingQuantity
			}
		}
	}

	if len(input.Cardholder) > 255 || len(input.Instructions) > 5000 {
		return ErrInvalidBillingQuantity
	}

	settings := model.BillingSettings{ID: 1, RateRial: rate, CardNumber: card, Cardholder: strings.TrimSpace(input.Cardholder), Instructions: input.Instructions, UpdatedBy: actor.AccountID, UpdatedAt: b.now().Unix()}

	return b.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&settings).Error; err != nil {
			return err
		}

		return billingAudit(tx, 0, actor.AccountID, nil, "payment_instructions_changed", "global rate="+input.RateRial)
	})
}

func billingRate(tx *gorm.DB, account *model.BillingAccount) (int64, error) {
	if account.RateRial != nil {
		return *account.RateRial, nil
	}

	var settings model.BillingSettings

	err := tx.First(&settings, 1).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DefaultPricePerGiB30DaysRial, nil
	}

	if err != nil {
		return 0, err
	}

	return settings.RateRial, nil
}
