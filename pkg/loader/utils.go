package loader

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/gomods/athens/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// check for path traversal and correct forward slashes
func validRelPath(p string) bool {
	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
		return false
	}

	return true
}

func untar(src, dst string) error {
	buffer, err := os.Open(src)
	if err != nil {
		return err
	}

	zr, err := gzip.NewReader(buffer)
	if err != nil {
		return err
	}

	tr := tar.NewReader(zr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		if !validRelPath(header.Name) {
			return fmt.Errorf("tar contained invalid name error %q\n", header.Name)
		}

		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {

		// if its a dir and it doesn't exist create it (with 0755 permission)
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		// if it's a file create it (with same permission)
		case tar.TypeReg:
			fileToWrite, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			// copy over contents
			if _, err := io.Copy(fileToWrite, tr); err != nil {
				return err
			}

			fileToWrite.Close()
		}
	}

	return nil
}

func unpackDumps(src, dst string) ([]string, error) {
	gzips, err := filepath.Glob(src + "athens_dump*.tar.gz")
	if err != nil {
		return nil, err
	}

	for _, g := range gzips {
		packedGzipDir, err := filepath.Abs(g)
		if err != nil {
			return nil, err
		}

		if err := untar(packedGzipDir, dst); err != nil {
			return nil, err
		}
	}

	return listDirs(dst)
}

func listDirs(rootSrc string) ([]string, error) {
	const op errors.Op = "utils.listDirs"
	var dirsList []string

	dirs, err := ioutil.ReadDir(rootSrc)
	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			return nil, errors.E(op, fmt.Errorf("%s is not a directory", dir.Name()))
		}

		dirsList = append(dirsList, dir.Name())
	}

	return dirsList, nil
}

func isAllFiles(path string) (bool, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return false, err
	}

	for _, f := range files {
		if f.IsDir() {
			return false, nil
		}
	}

	return true, nil
}
