package main

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

const (
	// TODO: keep updating that version
	version = "0.2.6"

	repoURL = "devusSs/steamquery"
)

func doSelfUpdate() error {
	writeInfo(fmt.Sprintf("Current version: %s\n", version))

	writeInfo("Checking for updates...")

	v := semver.MustParse(version)

	latest, err := selfupdate.UpdateSelf(v, repoURL)
	if err != nil {
		return err
	}

	if latest.Version.Equals(v) {
		writeSuccess("App is up to date")
	} else {
		writeSuccess(fmt.Sprintf("Successfully updated app to version %s", latest.Version.String()))
	}

	return nil
}
