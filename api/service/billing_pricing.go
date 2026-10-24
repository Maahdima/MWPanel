package service

import (
	"errors"
	"math/big"
	"strings"
)

const (
	BillingGiB                   int64 = 1 << 30
	DefaultPricePerGiB30DaysRial int64 = 100_000
	BillingPricingVersion              = "byte-rial-v1"
)

type BillingOperation string

const (
	BillingCreate       BillingOperation = "create"
	BillingAddTraffic   BillingOperation = "add_traffic"
	BillingExtendExpiry BillingOperation = "extend_expiry"
	BillingResetUsage   BillingOperation = "reset_usage"
	BillingFullReset    BillingOperation = "full_reset"
)

var (
	ErrInvalidBillingQuantity = errors.New("invalid billing quantity")
	ErrBillingOverflow        = errors.New("billing amount exceeds integer rial range")
)

// BillingPriceSnapshot is persisted with a charge, never recalculated using a later rate.
// Bytes means the configured quota for create/extend/full reset, the additional quota
// for add_traffic, and the actual consumed traffic for reset_usage.
// All intermediate arithmetic is exact. Nonnegative charges round half up to one
// whole rial exactly once, at posting. No floating point or rounded GiB is involved.
type BillingPriceSnapshot struct {
	Version    string           `json:"version"`
	Operation  BillingOperation `json:"operation"`
	Bytes      int64            `json:"bytes"`
	Days       int64            `json:"days"`
	RateRial   int64            `json:"rate_rial"`
	AmountRial int64            `json:"amount_rial"`
	Rounding   string           `json:"rounding"`
}

func PriceBillingOperation(operation BillingOperation, bytes, days, rateRial int64) (BillingPriceSnapshot, error) {
	var empty BillingPriceSnapshot
	if bytes < 0 || days < 0 || rateRial < 0 {
		return empty, ErrInvalidBillingQuantity
	}
	multiplier, divisor := int64(1), BillingGiB
	switch operation {
	case BillingCreate:
		if bytes == 0 || days == 0 {
			return empty, ErrInvalidBillingQuantity
		}
		multiplier = max(days, 30)
		divisor *= 30
	case BillingExtendExpiry:
		if bytes == 0 || days == 0 {
			return empty, ErrInvalidBillingQuantity
		}
	case BillingFullReset:
		if bytes == 0 || days == 0 {
			return empty, ErrInvalidBillingQuantity
		}
	case BillingAddTraffic:
		if bytes == 0 || days != 0 {
			return empty, ErrInvalidBillingQuantity
		}
	case BillingResetUsage:
		if days != 0 {
			return empty, ErrInvalidBillingQuantity
		}
	default:
		return empty, ErrInvalidBillingQuantity
	}
	numerator := new(big.Int).Mul(big.NewInt(bytes), big.NewInt(rateRial))
	numerator.Mul(numerator, big.NewInt(multiplier))
	denominator := big.NewInt(divisor)
	whole, remainder := new(big.Int), new(big.Int)
	whole.QuoRem(numerator, denominator, remainder)
	if remainder.Lsh(remainder, 1).Cmp(denominator) >= 0 {
		whole.Add(whole, big.NewInt(1))
	}
	if !whole.IsInt64() {
		return empty, ErrBillingOverflow
	}
	return BillingPriceSnapshot{
		Version: BillingPricingVersion, Operation: operation, Bytes: bytes,
		Days: days, RateRial: rateRial, AmountRial: whole.Int64(), Rounding: "half_up_rial",
	}, nil
}

// ParseBillingGiB accepts exact decimal GiB without using the legacy floating-point
// display conversion. Sub-byte quantities, exponents, signs and overflow are rejected.
func ParseBillingGiB(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, ErrInvalidBillingQuantity
	}
	dots, digits := 0, 0
	for _, c := range value {
		if c == '.' {
			dots++
			if dots > 1 {
				return 0, ErrInvalidBillingQuantity
			}
			continue
		}
		if c < '0' || c > '9' {
			return 0, ErrInvalidBillingQuantity
		}
		digits++
	}
	if digits == 0 || strings.HasPrefix(value, ".") || strings.HasSuffix(value, ".") || len(value) > 64 {
		return 0, ErrInvalidBillingQuantity
	}
	quantity, ok := new(big.Rat).SetString(value)
	if !ok {
		return 0, ErrInvalidBillingQuantity
	}
	quantity.Mul(quantity, new(big.Rat).SetInt64(BillingGiB))
	if !quantity.IsInt() || !quantity.Num().IsInt64() {
		return 0, ErrInvalidBillingQuantity
	}
	return quantity.Num().Int64(), nil
}
