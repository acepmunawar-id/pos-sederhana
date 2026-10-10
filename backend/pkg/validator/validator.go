// Package validator validates request DTOs and renders Indonesian error messages.
package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/posumkm/backend/pkg/apperror"
)

var v = newValidate()

func newValidate() *validator.Validate {
	val := validator.New(validator.WithRequiredStructEnabled())
	val.RegisterTagNameFunc(func(f reflect.StructField) string {
		for _, tag := range []string{"json", "form"} {
			name := strings.SplitN(f.Tag.Get(tag), ",", 2)[0]
			if name == "-" {
				return ""
			}
			if name != "" {
				return name
			}
		}
		return f.Name
	})
	return val
}

// Struct validates s and returns an *apperror.AppError with per-field messages, or nil.
func Struct(s any) error {
	err := v.Struct(s)
	if err == nil {
		return nil
	}
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return apperror.BadRequest("Format data tidak valid")
	}
	out := make([]apperror.FieldError, 0, len(verrs))
	for _, fe := range verrs {
		out = append(out, apperror.FieldError{Field: fieldPath(fe), Message: message(fe)})
	}
	return apperror.Validation(out)
}

// fieldPath strips the root struct name: "CheckoutRequest.items[0].quantity" → "items[0].quantity".
func fieldPath(fe validator.FieldError) string {
	ns := fe.Namespace()
	if i := strings.Index(ns, "."); i >= 0 {
		return ns[i+1:]
	}
	return fe.Field()
}

func message(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "min":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("minimal %s karakter", fe.Param())
		}
		if fe.Kind() == reflect.Slice {
			return fmt.Sprintf("minimal %s item", fe.Param())
		}
		return fmt.Sprintf("minimal %s", fe.Param())
	case "max":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("maksimal %s karakter", fe.Param())
		}
		return fmt.Sprintf("maksimal %s", fe.Param())
	case "gt":
		return fmt.Sprintf("harus lebih dari %s", fe.Param())
	case "gte":
		return fmt.Sprintf("minimal %s", fe.Param())
	case "lte":
		return fmt.Sprintf("maksimal %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("harus salah satu dari: %s", strings.ReplaceAll(fe.Param(), " ", ", "))
	case "nefield":
		return "tidak boleh sama dengan password lama"
	case "eqfield":
		return "konfirmasi tidak cocok"
	case "dive":
		return "isi tidak valid"
	}
	return "tidak valid"
}
