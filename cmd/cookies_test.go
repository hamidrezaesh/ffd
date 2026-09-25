package cmd

import (
	"fmt"
	"testing"
)

func TestParseCookieFile(t *testing.T) {
	cookies, err := parseCookieFile("../temp/example_cookies.txt")
	if err != nil {
		t.Fatal(err)
	}

	for _, cookie := range cookies {
		fmt.Printf(
			"Domain: %s\nCookie: %s\n\n",
			cookie.Domain,
			cookie.Cookie,
		)
	}
}
