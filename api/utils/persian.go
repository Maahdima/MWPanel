package utils

import (
	"strconv"
	"strings"
	"time"
)

// NormalizeDigits maps Persian and Arabic-Indic digits to ASCII digits.
func NormalizeDigits(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= '۰' && r <= '۹':
			b.WriteByte(byte('0' + r - '۰'))
		case r >= '٠' && r <= '٩':
			b.WriteByte(byte('0' + r - '٠'))
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ToPersianDigits maps ASCII digits to Persian digits.
func ToPersianDigits(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune('۰' + (r - '0'))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// FormatIntGrouped formats n with comma thousands separators.
func FormatIntGrouped(n int64) string {
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}

	raw := strconv.FormatInt(n, 10)
	if len(raw) <= 3 {
		return sign + raw
	}

	lead := len(raw) % 3
	if lead == 0 {
		lead = 3
	}
	var b strings.Builder
	b.WriteString(sign)
	b.WriteString(raw[:lead])
	for i := lead; i < len(raw); i += 3 {
		b.WriteByte(',')
		b.WriteString(raw[i : i+3])
	}
	return b.String()
}

// FormatTomans formats a toman amount with Persian digits.
func FormatTomans(n int64) string {
	grouped := strings.ReplaceAll(FormatIntGrouped(n), ",", "٬")
	return ToPersianDigits(grouped) + " تومان"
}

// FormatRials formats a rial amount with Persian digits.
func FormatRials(n int64) string {
	grouped := strings.ReplaceAll(FormatIntGrouped(n), ",", "٬")
	return ToPersianDigits(grouped) + " ریال"
}

// PersianPlanTitle is the shop label for a one-month plan, e.g. "۱ ماهه ۲ گیگ".
func PersianPlanTitle(trafficGB int) string {
	return ToPersianDigits("1") + " ماهه " + ToPersianDigits(strconv.Itoa(trafficGB)) + " گیگ"
}

// FormatPersianDate renders a stored YYYY-MM-DD date on the Persian calendar, e.g. "۲۲ آبان ۱۴۰۵".
func FormatPersianDate(raw string) string {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
	if err != nil {
		return strings.TrimSpace(raw)
	}
	year, month, day := gregorianToJalali(parsed.Year(), int(parsed.Month()), parsed.Day())
	months := []string{
		"", "فروردین", "اردیبهشت", "خرداد", "تیر", "مرداد", "شهریور",
		"مهر", "آبان", "آذر", "دی", "بهمن", "اسفند",
	}
	if month < 1 || month > 12 {
		return strings.TrimSpace(raw)
	}
	return ToPersianDigits(strconv.Itoa(day)) + " " + months[month] + " " + ToPersianDigits(strconv.Itoa(year))
}

func gregorianToJalali(gy, gm, gd int) (int, int, int) {
	gDaySums := [12]int{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334}
	gy2 := gy
	if gm > 2 {
		gy2++
	}
	days := 355666 + (365 * gy) + ((gy2 + 3) / 4) - ((gy2 + 99) / 100) + ((gy2 + 399) / 400) + gd + gDaySums[gm-1]
	jy := -1595 + (33 * (days / 12053))
	days %= 12053
	jy += 4 * (days / 1461)
	days %= 1461
	if days > 365 {
		jy += (days - 1) / 365
		days = (days - 1) % 365
	}
	var jm, jd int
	if days < 186 {
		jm = 1 + (days / 31)
		jd = 1 + (days % 31)
	} else {
		jm = 7 + ((days - 186) / 30)
		jd = 1 + ((days - 186) % 30)
	}
	return jy, jm, jd
}
func FormatCardNumber(raw string) string {
	digits := CardDigits(raw)
	if digits == "" {
		return strings.TrimSpace(raw)
	}
	groups := make([]string, 0, (len(digits)+3)/4)
	for len(digits) > 4 {
		groups = append(groups, digits[:4])
		digits = digits[4:]
	}
	groups = append(groups, digits)
	return strings.Join(groups, "-")
}

// CardDigits returns the card as ASCII digits, with no separators.
func CardDigits(raw string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, NormalizeDigits(raw))
}
