package timehelper

import (
	"fmt"
	"time"
)

func ParseTime(input string) (int, error) {
	t, err := time.Parse("15:04:05", input)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	return t.Second(), nil
}
