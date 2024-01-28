package testkit

import (
	"fmt"

	"github.com/alvii147/flagger-api/pkg/utils"
)

// MustGenerateRandomString generates a random string and panics on error.
func MustGenerateRandomString(
	length int,
	allowLowerAlpha bool,
	allowUpperAlpha bool,
	allowNumeric bool,
) string {
	s, err := utils.GenerateRandomString(length, allowLowerAlpha, allowUpperAlpha, allowNumeric)
	if err != nil {
		panic(fmt.Sprintf("MustGenerateRandomString failed to utils.GenerateRandomString: %v", err))
	}

	return s
}

// GenerateFakeEmail generates a randomized email addresss.
func GenerateFakeEmail() string {
	return fmt.Sprintf(
		"%s@%s.%s",
		MustGenerateRandomString(12, true, false, true),
		MustGenerateRandomString(10, true, false, false),
		MustGenerateRandomString(3, true, false, false),
	)
}

// GenerateFakeEmail generates a randomized password.
func GenerateFakePassword() string {
	password := MustGenerateRandomString(20, true, true, true)

	return password
}
