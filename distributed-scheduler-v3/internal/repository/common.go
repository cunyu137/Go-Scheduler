package repository

import (
	"errors"
	"strings"
)

func IsDuplicateErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "1062")
}

func IgnoreNotFound(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	return err
}

var ErrNotFound = errors.New("not found")
