package util

import (
	"errors"
	"regexp"
	"strconv"
)

var sizeRegex = regexp.MustCompile(`^([1-9][0-9]*){1}(MB|GB){1}$`)

func ParseSize(v string) (int64, error) {
	submatches := sizeRegex.FindStringSubmatch(v)
	if len(submatches) != 3 {
		return 0, errors.New("invalid size")
	}
	n, err := strconv.ParseInt(submatches[1], 10, 0)
	if err != nil {
		return 0, err
	}
	var b int
	switch submatches[2] {
	case "MB":
		b = 1048576
	case "GB":
		b = 1073741824
	default:
		return 0, errors.New("invalid size")
	}
	return int64(b) * n, nil
}
