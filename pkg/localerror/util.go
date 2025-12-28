package localerror

import (
	"base-be-golang/internal/constant"
	"errors"
	"gorm.io/gorm"
)

type InvalidDataError struct {
	Msg             string
	DataToTemplated map[string]string
}

func (e InvalidDataError) Error() string {
	return e.Msg
}

type AccessControlError struct {
	Msg string
}

func (e AccessControlError) Error() string {
	return e.Msg
}

func AccessNotAllowedUserNotFound(err error) error {
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return AccessControlError{Msg: constant.SessionExpired}
	}

	return err
}

func IsNotFound(target error, source error) bool {
	return errors.Is(source, InvalidDataError{Msg: target.Error()})
}

func NotFound(err error, msg string) error {
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return InvalidDataError{Msg: msg}
	}
	return err
}

func IsNotFoundStr(target string, source error) bool {
	var newErr = InvalidDataError{}
	if errors.As(source, &newErr) {
		return target == newErr.Msg
	}

	return false
}

func InvalidData(msg string) error {
	return InvalidDataError{Msg: msg}
}

func InvalidDataWithData(msg string, data map[string]string) error {
	return InvalidDataError{Msg: msg, DataToTemplated: data}
}
