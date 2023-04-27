package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/blang/semver"
	"github.com/google/go-github/v30/github"
)

const (
	// TODO: keep updating that version
	version = "0.2.7"

	repoOwner = "devusSs"
	repoName  = "steamquery"
)

var (
	client *github.Client = github.NewClient(nil)
)

func checkLatestReleaseGithub() (*github.RepositoryRelease, error) {
	rel, res, err := client.Repositories.ListReleases(context.Background(), repoOwner, repoName, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("got unwanted Github error code: %d", res.StatusCode)
	}

	latestRelease := rel[0]

	return latestRelease, nil
}

func checkVersionMatch(release *github.RepositoryRelease) (bool, error) {
	currentVersion, err := semver.Parse(strings.ReplaceAll(version, "v", ""))
	if err != nil {
		return false, err
	}

	latestVersion, err := semver.Parse(strings.ReplaceAll(*release.TagName, "v", ""))
	if err != nil {
		return false, err
	}

	return latestVersion.Equals(currentVersion), nil
}

// Finds the proper release file and download url for our OS and platform.
func findMatchingOSAndPlatform(release *github.RepositoryRelease) (string, error) {
	osV := runtime.GOOS
	pV := runtime.GOARCH

	for _, asset := range release.Assets {
		name := strings.ToLower(*asset.Name)

		if strings.Contains(name, osV) && strings.Contains(name, pV) {
			return *asset.BrowserDownloadURL, nil
		}
	}

	return "", errors.New("no matching release found")
}

// Helper function to unzip the tar.gz file we got from Github.
// Original answer provided by ChatGPT, adapted to my needs.
func handlePatchDownloadAndUnzip(url string) error {
	// Create a temporary
	if _, err := os.Stat("./tmp"); os.IsNotExist(err) {
		if err := os.Mkdir("./tmp", os.ModePerm); err != nil {
			return err
		}
	}

	// Create a temporary file to save the downloaded archive
	out, err := os.CreateTemp("./tmp", "download-*.tar.gz")
	if err != nil {
		return err
	}
	defer out.Close()

	// Download the file
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the downloaded data to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	// Unzip the file
	archive, err := os.Open(out.Name())
	if err != nil {
		return err
	}
	defer archive.Close()

	gzipReader, err := gzip.NewReader(archive)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Extract the file
		path := filepath.Join(".", header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			os.MkdirAll(path, info.Mode())
		} else {
			outFile, err := os.Create(path)
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, tarReader)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
