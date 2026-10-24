package service

import (
	"errors"
	"math"
	"testing"
)

func TestBillingPricingRules(t *testing.T) {
	cases := []struct {
		name                    string
		op                      BillingOperation
		bytes, days, rate, want int64
	}{
		{"create 15 days", BillingCreate, 10 * BillingGiB, 15, 100000, 1000000},
		{"create 30 days", BillingCreate, 10 * BillingGiB, 30, 100000, 1000000},
		{"create 45 days", BillingCreate, 10 * BillingGiB, 45, 100000, 1500000},
		{"create 60 days", BillingCreate, 10 * BillingGiB, 60, 100000, 2000000},
		{"add 5 GiB", BillingAddTraffic, 5 * BillingGiB, 0, 100000, 500000},
		{"extend 15 days", BillingExtendExpiry, 10 * BillingGiB, 15, 100000, 1000000},
		{"extend 30 days", BillingExtendExpiry, 10 * BillingGiB, 30, 100000, 1000000},
		{"extend 60 days", BillingExtendExpiry, 10 * BillingGiB, 60, 100000, 1000000},
		{"usage reset 4 GiB", BillingResetUsage, 4 * BillingGiB, 0, 100000, 400000},
		{"full reset quota 100 GiB", BillingFullReset, 100 * BillingGiB, 60, 100000, 10000000},
		{"empty usage reset", BillingResetUsage, 0, 0, 100000, 0},
		{"fractional usage", BillingResetUsage, BillingGiB / 4, 0, 100000, 25000},
		{"below half rial", BillingResetUsage, BillingGiB/2 - 1, 0, 1, 0},
		{"exact half rial", BillingResetUsage, BillingGiB / 2, 0, 1, 1},
		{"above half rial", BillingResetUsage, BillingGiB/2 + 1, 0, 1, 1},
		{"zero override", BillingCreate, BillingGiB, 30, 0, 0},
		{"large exact intermediate", BillingCreate, BillingGiB, 30, math.MaxInt64, math.MaxInt64},
		{"individual rate", BillingAddTraffic, 5 * BillingGiB, 0, 123457, 617285},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := PriceBillingOperation(tc.op, tc.bytes, tc.days, tc.rate)
			if err != nil {
				t.Fatal(err)
			}
			if got.AmountRial != tc.want {
				t.Fatalf("got %d rials, want %d", got.AmountRial, tc.want)
			}
			if got.Version != BillingPricingVersion || got.Bytes != tc.bytes || got.Days != tc.days || got.RateRial != tc.rate || got.Operation != tc.op {
				t.Fatalf("incomplete snapshot: %+v", got)
			}
		})
	}
}

func TestBillingPricingRejectsInvalidQuantities(t *testing.T) {
	for _, tc := range []struct {
		op                BillingOperation
		bytes, days, rate int64
	}{
		{BillingCreate, 0, 30, 100000}, {BillingCreate, BillingGiB, 0, 100000},
		{BillingCreate, -1, 30, 100000}, {BillingCreate, BillingGiB, 30, -1},
		{BillingExtendExpiry, BillingGiB, 0, 100000}, {BillingFullReset, 0, 30, 100000},
		{BillingAddTraffic, 0, 0, 100000}, {BillingAddTraffic, BillingGiB, 30, 100000},
		{BillingResetUsage, -1, 0, 100000}, {BillingResetUsage, 1, 1, 100000},
		{BillingOperation("unknown"), BillingGiB, 30, 100000},
	} {
		if _, err := PriceBillingOperation(tc.op, tc.bytes, tc.days, tc.rate); !errors.Is(err, ErrInvalidBillingQuantity) {
			t.Errorf("%+v: got %v", tc, err)
		}
	}
	if _, err := PriceBillingOperation(BillingCreate, math.MaxInt64, math.MaxInt64, math.MaxInt64); !errors.Is(err, ErrBillingOverflow) {
		t.Fatalf("want overflow, got %v", err)
	}
}

func TestParseBillingGiB(t *testing.T) {
	for value, want := range map[string]int64{"0": 0, "1": BillingGiB, "1.25": BillingGiB + BillingGiB/4, " 0.5 ": BillingGiB / 2, "0.000000000931322574615478515625": 1} {
		got, err := ParseBillingGiB(value)
		if err != nil || got != want {
			t.Errorf("%q: %d %v, want %d", value, got, err, want)
		}
	}
	for _, value := range []string{"", "-1", "+1", "1e3", "NaN", "1/2", ".5", "1.", "0.0000000001", "99999999999999999999999999", "1.2.3"} {
		if _, err := ParseBillingGiB(value); err == nil {
			t.Errorf("accepted %q", value)
		}
	}
}
