package config

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	validatorOnce sync.Once
	vl            *validator.Validate
)
