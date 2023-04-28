package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/blang/semver"
	"github.com/google/go-github/v30/github"
)

const (
	// TODO: keep updating that version
	version = "0.2.8"

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
	// Create a temporary directory.
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
			if err := os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
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

// Helper function to unzip the .zip file we got from Github.
// Original answer provided by ChatGPT, adapted to my needs.
func handlePatchDownloadAndUnzipWindows(url string) error {
	// Create a temporary directory.
	if _, err := os.Stat("./tmp"); os.IsNotExist(err) {
		if err := os.Mkdir("./tmp", os.ModePerm); err != nil {
			return err
		}
	}

	// Name of the local file to save the .zip as
	filename := "./tmp/latest.zip"

	// Create the file on disk to save the .zip to
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	// Download the .zip from the URL and save it to the file on disk
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	// Unzip the file
	zipReader, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	var binaryPath string
	for _, file := range zipReader.File {
		if filepath.Ext(file.Name) == ".exe" {
			path := filepath.Join("./tmp", file.Name)

			fileReader, err := file.Open()
			if err != nil {
				return err
			}
			defer fileReader.Close()

			targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				return err
			}
			defer targetFile.Close()

			if _, err := io.Copy(targetFile, fileReader); err != nil {
				return err
			}

			binaryPath = path
			break
		}
	}

	if binaryPath == "" {
		return errors.New("could not find binary in .zip file")
	}

	// Replace the current program binary with the binary from the .zip
	currentBinaryPath, err := os.Executable()
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		// On Windows, we need to close the current program before replacing its binary
		if err := exec.Command("taskkill", "/F", "/IM", filepath.Base(currentBinaryPath)).Run(); err != nil {
			return err
		}
	}

	err = os.Rename(binaryPath, currentBinaryPath)
	if err != nil {
		return err
	}

	return nil
}
