package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

const (
	updateURL = "https://api.github.com/repos/devusSs/steamquery/releases/latest"
)

var (
	buildVersion = ""
	buildDate    = ""
	buildOS      = runtime.GOOS
	buildArch    = runtime.GOARCH
	goVersion    = runtime.Version()
)

// Function to print build information.
func printBuildInformation() {
	log.Printf("[%s] Build version: \t\t%s\n", infSign, buildVersion)
	log.Printf("[%s] Build date: \t\t%s\n", infSign, buildDate)
	log.Printf("[%s] Build OS: \t\t%s\n", infSign, buildOS)
	log.Printf("[%s] Build arch: \t\t%s\n", infSign, buildArch)
	log.Printf("[%s] Go version: \t\t%s\n", infSign, goVersion)
}

// Queries the latest release from Github repo.
func findLatestReleaseURL() (string, string, string, error) {
	resp, err := http.Get(updateURL)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	var release githubRelease

	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return "", "", "", err
	}

	// Fix versions / architecture to match Github releases.
	if buildArch == "amd64" {
		buildArch = "x86_64"
	}

	if buildArch == "386" {
		buildArch = "i386"
	}

	// Find matching release for our OS & architecture.
	for _, asset := range release.Assets {
		releaseName := strings.ToLower(asset.Name)

		if strings.Contains(releaseName, buildArch) && strings.Contains(releaseName, buildOS) {
			// Format the changelog body accordingly.
			changeSplit := strings.Split(strings.ReplaceAll(strings.TrimSpace(release.Body), "## Changelog", ""), "\n")
			for i, line := range changeSplit {
				changeSplit[i] = strings.ReplaceAll(fmt.Sprintf("\t\t\t%s", line), "*", "-")
			}
			changelog := strings.Join(changeSplit, "\n")
			return asset.BrowserDownloadURL, release.TagName, changelog, nil
		}
	}

	return "", "", "", errors.New("no matching release found")
}

// Compare current version with latest version
func newerVersionAvailable(newVersion string) (bool, error) {
	currentBuild := strings.ReplaceAll(buildVersion, "v", "")
	newBuild := strings.ReplaceAll(newVersion, "v", "")

	vOld, err := semver.NewVersion(currentBuild)
	if err != nil {
		return false, err
	}

	vNew, err := semver.NewVersion(newBuild)
	if err != nil {
		return false, err
	}

	return !vNew.Equal(vOld), nil
}

// Perform the actual patch.
func doUpdate(url string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	if err := selfupdate.UpdateTo(url, exe); err != nil {
		return err
	}

	return nil
}

func periodicUpdateCheck() error {
	_, versionCheck, _, err := findLatestReleaseURL()
	if err != nil {
		return err
	}

	newVersionAvailable, err := newerVersionAvailable(versionCheck)
	if err != nil {
		return err
	}

	if newVersionAvailable {
		log.Printf("[%s] New version available (%s). Please restart your app soon\n", warnSign, versionCheck)
	}

	return nil
}
