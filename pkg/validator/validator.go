// Package validator membungkus go-playground/validator/v10 supaya bisa
// dipasang sebagai e.Validator di Echo (echo.Context.Validate() manggil ini).
// Dipakai bareng semua module (product, cart, order, dst) — didaftarkan
// sekali di bootstrap/app.go.
package validator

import (
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	return &Validator{validate: validator.New()}
}

// Validate mengimplementasikan echo.Validator interface.
func (v *Validator) Validate(i interface{}) error {
	return v.validate.Struct(i)
}
