package archive

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Pack compresses the source directory into the writer using tar.
// It preserves file metadata.
func Pack(src string, w io.Writer) error {
	tw := tar.NewWriter(w)
	defer tw.Close()

	// Ensure src is absolute or clean
	src, err := filepath.Abs(src)
	if err != nil {
		return err
	}

	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Dir(src)
	} else {
		// If it's a single file, just archive it at root? 
		// The prompt says "encrypt a directory". 
		// But let's handle file too just in case.
		baseDir = filepath.Dir(src)
	}

	return filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create tar header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// Update name to be relative to baseDir
		relPath, err := filepath.Rel(baseDir, file)
		if err != nil {
			return err
		}
		
		// Ensure forward slashes for tar
		header.Name = filepath.ToSlash(relPath)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		return nil
	})
}

// Unpack extracts the tar stream from r into dstDir.
func Unpack(r io.Reader, dstDir string) error {
	dstDir, err := filepath.Abs(dstDir)
	if err != nil {
		return err
	}

	tr := tar.NewReader(r)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Sanitize path to prevent Zip Slip
		target := filepath.Join(dstDir, header.Name)
		if !strings.HasPrefix(target, filepath.Clean(dstDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			dir := filepath.Dir(target)
			if err := os.MkdirAll(dir, 0755); err != nil {
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
