package config

import "errors"

var (
	// ErrMissingLTIConfig is returned when LTI configuration is missing in production
	ErrMissingLTIConfig = errors.New("LTI configuration required in production mode")

	// ErrInsecureSessionSecret is returned when using default session secret in production
	ErrInsecureSessionSecret = errors.New("session secret must be changed in production")
)
