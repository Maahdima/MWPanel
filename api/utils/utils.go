package utils

import (
	"fmt"
	"math/rand"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/gommon/log"
)

func Ptr(s string) *string { return &s }

func DerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ParseStringToInt(s string) int64 {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

func HandshakeStatus(disabled string, lastHandshake *string, threshold time.Duration) (age time.Duration, online bool, err error) {
	if lastHandshake == nil || strings.TrimSpace(*lastHandshake) == "" {
		return 0, false, nil
	}

	age, err = ParseCustomDuration(*lastHandshake)
	if err != nil {
		return 0, false, err
	}

	if strings.EqualFold(strings.TrimSpace(disabled), "false") && age < threshold {
		return age, true, nil
	}

	return age, false, nil
}

func FormatDataSize(b int64) string {
	if b < 0 {
		b = 0
	}

	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)

	switch {
	case b >= gb:
		return fmt.Sprintf("%.2f GB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func FormatPrettyDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	d = d.Truncate(time.Second)

	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute
	d -= minutes * time.Minute
	seconds := d / time.Second

	hms := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	if days > 0 {
		return fmt.Sprintf("%dd %s", days, hms)
	}
	return hms
}

func ParseCustomDuration(s string) (time.Duration, error) {
	re := regexp.MustCompile(`(\d+)([wdhms])`)
	matches := re.FindAllStringSubmatch(s, -1)

	var total time.Duration
	for _, match := range matches {
		num, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, err
		}
		switch match[2] {
		case "w":
			total += time.Duration(num) * 7 * 24 * time.Hour
		case "d":
			total += time.Duration(num) * 24 * time.Hour
		case "h":
			total += time.Duration(num) * time.Hour
		case "m":
			total += time.Duration(num) * time.Minute
		case "s":
			total += time.Duration(num) * time.Second
		default:
			return 0, fmt.Errorf("unknown duration unit: %s", match[2])
		}
	}
	return total, nil
}

func BytesToGB(b int64) string {
	return fmt.Sprintf("%.1f", float64(b)/float64(1024*1024*1024))
}

func GBToBytes(s string) int64 {
	gb, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int64(gb * 1024 * 1024 * 1024)
}

func IsPeerExpired(expireTime *string, now time.Time) bool {
	if expireTime == nil {
		return false
	}

	raw := strings.TrimSpace(*expireTime)
	if raw == "" {
		return false
	}

	expireDay, err := time.ParseInLocation("2006-01-02", raw, now.Location())
	if err != nil {
		return false
	}

	year, month, day := now.Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
	return !today.Before(expireDay)
}

func IsPeerSharable(isShared bool, shareExpireTime *string) bool {
	if !isShared {
		log.Errorf("peer is not shared")
		return false
	}

	if shareExpireTime != nil && strings.TrimSpace(*shareExpireTime) != "" {
		if _, err := time.Parse("2006-01-02", strings.TrimSpace(*shareExpireTime)); err != nil {
			log.Errorf("failed to parse share expire time")
			return false
		}
		if IsPeerExpired(shareExpireTime, time.Now()) {
			log.Errorf("share link has expired")
			return false
		}
	}

	return true
}

func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func IPToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	if ip == nil {
		return 0
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

func FormatDuration(d time.Duration) string {
	d = d.Truncate(time.Second)

	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
