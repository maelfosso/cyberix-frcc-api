package models

import (
	"regexp"
)

// emailAddressMatcher for valid email addresses.
// See https://regex101.com/r/1BEPJo/latest for an interactive breakdown of the regexp.
// See https://html.spec.whatwg.org/#valid-e-mail-address for the definition.
var emailAddressMatcher = regexp.MustCompile(
	// Start of string
	`^` +
		// Local part of the address. Note that \x60 is a backtick (`) character.
		`(?P<local>[a-zA-Z0-9.!#$%&'*+/=?^_\x60{|}~-]+)` +
		`@` +
		// Domain of the address
		`(?P<domain>[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*)` +
		// End of string
		`$`,
)

type Email string

func (e Email) IsValid() bool {
	return emailAddressMatcher.MatchString(string(e))
}

func (e Email) String() string {
	return string(e)
}
