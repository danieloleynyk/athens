package dumper

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/spf13/afero"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func compress(src, dst string) error {
	var buf bytes.Buffer

	Fs := afero.NewOsFs()
	zr := gzip.NewWriter(io.Writer(&buf))
	tw := tar.NewWriter(zr)

	// walk through every file in the folder
	afero.Walk(Fs, src, func(file string, fi os.FileInfo, err error) error {

		// generate tar header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(strings.TrimPrefix(file, "/tmp/"))

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})

	// produce tar
	if err := tw.Close(); err != nil {
		return err
	}
	// produce gzip
	if err := zr.Close(); err != nil {
		return err
	}

	fileName := fmt.Sprintf("%s.tar.gz", strings.TrimPrefix(src, "/tmp/"))
	dst, err := filepath.Abs(path.Join(dst, fileName))
	if err != nil {
		return err
	}

	fileToWrite, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, os.FileMode(0655))
	if err != nil {
		return err
	}

	_, err = io.Copy(fileToWrite, &buf)
	return err
}
