package utils

import (
	"fmt"
	"strconv"
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

func ParseStringToFloat(s string) float64 {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
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
