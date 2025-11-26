package runner

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/v66/github"
	"github.com/sirupsen/logrus"
	gh "github.com/skevetter/bladerunner/internal/github"
)

type Installer struct {
	client *gh.Client
}

func NewInstaller(client *gh.Client) *Installer {
	return &Installer{client: client}
}

// Install downloads and installs the runner binary
func (i *Installer) Install(ctx context.Context, owner, repo, destDir string) error {
	logrus.Infof("Finding runner binary for %s/%s (%s/%s)", owner, repo, runtime.GOOS, runtime.GOARCH)

	downloads, err := i.client.ListRunnerApplicationDownloads(owner, repo)
	if err != nil {
		return fmt.Errorf("failed to list runner downloads: %w", err)
	}

	download := i.findMatchingDownload(downloads)
	if download == nil {
		return fmt.Errorf("no matching runner binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	logrus.Infof("Downloading runner from %s", *download.DownloadURL)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Download to temp file
	tmpFile, err := os.CreateTemp("", "runner-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if err := i.downloadFile(ctx, *download.DownloadURL, tmpFile); err != nil {
		return err
	}

	// Verify checksum if available
	if download.SHA256Checksum != nil && *download.SHA256Checksum != "" {
		logrus.Info("Verifying checksum...")
		if err := i.verifyChecksum(tmpFile.Name(), *download.SHA256Checksum); err != nil {
			return err
		}
	}

	// Extract
	logrus.Info("Extracting runner...")
	if err := i.extract(tmpFile.Name(), destDir); err != nil {
		return fmt.Errorf("failed to extract runner: %w", err)
	}

	logrus.Info("Runner installed successfully")
	return nil
}

func (i *Installer) findMatchingDownload(downloads []*github.RunnerApplicationDownload) *github.RunnerApplicationDownload {
	goOS := runtime.GOOS
	goArch := runtime.GOARCH

	// Map Go OS/Arch to GitHub Runner OS/Arch
	var targetOS, targetArch string

	switch goOS {
	case "linux":
		targetOS = "linux"
	case "darwin":
		targetOS = "osx" // GitHub uses 'osx' for macOS
	case "windows":
		targetOS = "win"
	default:
		return nil
	}

	switch goArch {
	case "amd64":
		targetArch = "x64"
	case "arm64":
		targetArch = "arm64"
	case "arm":
		targetArch = "arm"
	default:
		return nil
	}

	for _, d := range downloads {
		if d.OS == nil || d.Architecture == nil {
			continue
		}
		if *d.OS == targetOS && *d.Architecture == targetArch {
			return d
		}
	}

	return nil
}

func (i *Installer) downloadFile(ctx context.Context, url string, dest *os.File) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	_, err = io.Copy(dest, resp.Body)
	return err
}

func (i *Installer) verifyChecksum(filepath, expectedChecksum string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return err
	}

	actualChecksum := hex.EncodeToString(hasher.Sum(nil))
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

func (i *Installer) extract(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)

		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", target)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}

	return nil
}
