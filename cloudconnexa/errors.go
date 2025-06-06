package cloudconnexa

import "errors"

// ErrCredentialsRequired is returned when client ID or client secret is missing.
var ErrCredentialsRequired = errors.New("both client_id and client_secret credentials must be specified")
