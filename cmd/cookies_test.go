package cmd

import (
	"fmt"
	"os"
	"testing"
)

func TestParseCookieFile(t *testing.T) {
	path := "../temp/example_cookies.txt"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("cookie file does not exist")
	}

	cookies, err := parseCookieFile(path)
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
