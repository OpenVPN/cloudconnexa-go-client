package cloudconnexa

import "errors"

// ErrCredentialsRequired is returned when client ID or client secret is missing.
var ErrCredentialsRequired = errors.New("both client_id and client_secret credentials must be specified")

// ErrEmptyID is returned when an empty ID is provided to a method that requires one.
var ErrEmptyID = errors.New("id cannot be empty")
