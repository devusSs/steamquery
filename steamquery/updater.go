package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
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

// ChatGPT answer, adapted for own needs.
func findLatestReleaseURL() (string, string, error) {
	resp, err := http.Get(updateURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name        string `json:"name"`
			DownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return "", "", err
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
			return asset.DownloadURL, release.TagName, nil
		}
	}

	return "", "", errors.New("no matching release found")
}

// Compare current version with latest version
func newerVersionAvailable(newVersion string) (bool, error) {
	vOld, err := semver.NewVersion(buildVersion)
	if err != nil {
		return false, err
	}

	vNew, err := semver.NewVersion(strings.ReplaceAll(newVersion, "v", ""))
	if err != nil {
		return false, err
	}

	return !vNew.Equal(vOld), nil
}

// Function to download and unzip as well as install new version on Windows.
// Original answer by ChatGPT, adapted to own needs.
func patchWindows(url string) error {
	// Download the zip file
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create a temporary file to store the downloaded zip file
	tempZipFile, err := os.CreateTemp("", "temp-zip-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tempZipFile.Name())
	defer tempZipFile.Close()

	// Write the zip file contents to the temporary file
	_, err = io.Copy(tempZipFile, resp.Body)
	if err != nil {
		return err
	}

	// Open the zip file
	r, err := zip.OpenReader(tempZipFile.Name())
	if err != nil {
		return err
	}
	defer r.Close()

	// Find the executable file in the zip
	var exeFile *zip.File
	for _, f := range r.File {
		if filepath.Ext(f.Name) == ".exe" {
			exeFile = f
			break
		}
	}
	if exeFile == nil {
		return errors.New("could not find executable file in zip")
	}

	// Extract the executable file to the current directory
	exePath := filepath.Join(".", filepath.Base(exeFile.Name))
	exeFileReader, err := exeFile.Open()
	if err != nil {
		return err
	}
	defer exeFileReader.Close()

	exeFileWriter, err := os.Create(exePath)
	if err != nil {
		return err
	}
	defer exeFileWriter.Close()
	_, err = io.Copy(exeFileWriter, exeFileReader)
	if err != nil {
		return err
	}

	// TODO: this crashes program (? not tested yet)
	// Replace the current executable with the extracted executable
	err = os.Rename(exePath, os.Args[0])
	if err != nil {
		return err
	}

	return nil
}

// Function to download and unzip as well as install new version on Unix systems.
// Original answer by ChatGPT, adapted to own needs.
func patchUnix(url string) error {
	// Download the .tar.gz file
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Unzip the file
	gzipReader, err := gzip.NewReader(resp.Body)
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

		target := filepath.Join(".", header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			file, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(file, tarReader); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown file type: %v", header.Typeflag)
		}
	}

	// Replace the current program with the unzipped program
	selfPath, err := os.Executable()
	if err != nil {
		return err
	}

	// TODO: this crashes program
	if err := os.Rename(filepath.Join(".", "program"), selfPath); err != nil {
		return err
	}

	return nil
}
