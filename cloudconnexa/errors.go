package cloudconnexa

import "errors"

// ErrCredentialsRequired is returned when client ID or client secret is missing.
var ErrCredentialsRequired = errors.New("both client_id and client_secret credentials must be specified")

// ErrEmptyID is returned when an empty ID is provided to a method that requires one.
var ErrEmptyID = errors.New("id cannot be empty")

// ErrResponseTooLarge is returned when a response body exceeds the configured size limit.
var ErrResponseTooLarge = errors.New("response body exceeds maximum allowed size")

// ErrInvalidBaseURL is returned when the base URL is malformed or cannot be parsed.
var ErrInvalidBaseURL = errors.New("invalid base URL")

// ErrHTTPSRequired is returned when HTTP is used but HTTPS is required for security.
var ErrHTTPSRequired = errors.New("HTTPS required: HTTP is not allowed for OAuth credentials")
