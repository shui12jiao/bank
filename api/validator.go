package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/shui12jiao/my_simplebank/util"
)

var (
	validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
		if currency, ok := fieldLevel.Field().Interface().(string); ok {
			return util.IsSupportedCurrency(currency)
		}
		return false
	}

	validUsername validator.Func = func(fieldLevel validator.FieldLevel) bool {
		if username, ok := fieldLevel.Field().Interface().(string); ok {
			if err := util.ValidateUsername(username); err == nil {
				return true
			}
		}
		return false
	}

	validPassword validator.Func = func(fieldLevel validator.FieldLevel) bool {
		if password, ok := fieldLevel.Field().Interface().(string); ok {
			if err := util.ValidatePassword(password); err == nil {
				return true
			}
		}
		return false
	}

	validFullName validator.Func = func(fieldLevel validator.FieldLevel) bool {
		if fullName, ok := fieldLevel.Field().Interface().(string); ok {
			if err := util.ValidateFullName(fullName); err == nil {
				return true
			}
		}
		return false
	}

	validEmail validator.Func = func(fieldLevel validator.FieldLevel) bool {
		if email, ok := fieldLevel.Field().Interface().(string); ok {
			if err := util.ValidateEmail(email); err == nil {
				return true
			}
		}
		return false
	}
)
