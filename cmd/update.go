package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"

	"github.com/hamidrezaesh/ffd/internal/version"
	"github.com/spf13/cobra"
)

type Release struct {
	TagName string `json:"tag_name"`
}

func getLatestVersion() (string, error) {
	url := "https://api.github.com/repos/hamidrezaesh/ffd/releases/latest"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var release Release

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update ffd",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// detecting os
			log.Print("Detecting OS...")
			os := runtime.GOOS
			log.Printf("OS: %v\n", os)

			// checking version
			latestVersion, err := getLatestVersion()
			latestVersion = strings.TrimPrefix(latestVersion, "v")
			if err != nil {
				log.Fatal(err)
			}
			if latestVersion == version.Version {
				log.Printf("ffd is up to date (%v)\n", version.Version)
				return
			}

			log.Printf("updating to the latest version (%v)...\n", latestVersion)

			var cmd *exec.Cmd

			if os == "windows" {
				cmd = exec.Command(
					"powershell",
					"-Command",
					"Set-ExecutionPolicy -Scope CurrentUser RemoteSigned; irm https://raw.githubusercontent.com/hamidrezaesh/ffd/main/scripts/install.ps1 | iex",
				)
			} else if os == "linux" || os == "darwin" {
				cmd = exec.Command(
					"sh",
					"-c",
					"curl -fsSL https://raw.githubusercontent.com/hamidrezaesh/ffd/main/scripts/install.sh | sh",
				)
			} else {
				log.Fatalf("Unsupported OS: %v", runtime.GOOS)
			}

			err = cmd.Run()
			if err != nil {
				log.Fatalf("Error while updating ffd: %v\n", err)
			}
		}
	},
}
