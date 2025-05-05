package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/starjardin/simplebank/utils"
)

var valiCurrency validator.Func = func(fl validator.FieldLevel) bool {
	currency, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	return utils.IsSupportedCurrency(currency)
}
