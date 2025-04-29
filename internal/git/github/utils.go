package github

import (
	"fmt"
	"regexp"
)

var valid = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func ValidateOrgName(orgName string) error {
	if !valid.MatchString(orgName) {
		return fmt.Errorf("invalid org name: %s. Must start with a letter or underscore and contain only letters, numbers, and underscores", orgName)
	}
	return nil
}
