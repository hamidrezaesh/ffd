//go:build windows

package disk

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

/*
Because of windows' architecture, we cant just make a file and allocate
like 100MB-200MB of file and forget about 0-100 part.
So instead we just make a temp file for each chunk, then finally
we combine the temp chunks into a file. and also remove the temp files.
*/

func AllocateChunk(path string, size int64) (*os.File, error) {
	dir := filepath.Dir(path)

	// Check free disk space.
	free, err := FreeSpace(dir)
	if err != nil {
		return nil, err
	}

	if free < size {
		return nil, errors.New("not enough disk space")
	}

	// Create the directory.
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Hide the directory.
	if err := windows.SetFileAttributes(
		windows.StringToUTF16Ptr(dir),
		windows.FILE_ATTRIBUTE_HIDDEN,
	); err != nil {
		return nil, err
	}

	// Create the chunk file.
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	// Hide the chunk file.
	if err := windows.SetFileAttributes(
		windows.StringToUTF16Ptr(path),
		windows.FILE_ATTRIBUTE_HIDDEN,
	); err != nil {
		file.Close()
		return nil, err
	}

	return file, nil
}

// Append appends the contents of src to dst and removes src
// after the contents have been successfully copied.
func Append(dst, src string) error {
	out, err := os.OpenFile(
		dst,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		0644,
	)
	if err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		out.Close()
		return err
	}

	_, copyErr := io.Copy(out, in)

	closeInErr := in.Close()
	closeOutErr := out.Close()

	if copyErr != nil {
		return copyErr
	}

	if closeInErr != nil {
		return closeInErr
	}

	if closeOutErr != nil {
		return closeOutErr
	}

	return os.Remove(src)
}
