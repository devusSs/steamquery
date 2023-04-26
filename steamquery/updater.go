package main

import (
	"log"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

const (
	// TODO: keep updating that version
	version = "0.2.5"

	repoURL = "devusSs/steamquery"
)

func doSelfUpdate() error {
	log.Println("[INFO] Checking for updates...")

	v := semver.MustParse(version)
	latest, err := selfupdate.UpdateSelf(v, repoURL)
	if err != nil {
		return err
	}

	if latest.Version.Equals(v) {
		log.Printf("[INFO] App is up to date")
	} else {
		log.Printf("[SUCCESS] Successfully updated app to version %s\n", latest.Version.String())
	}

	return nil
}
