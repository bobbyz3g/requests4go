package requests4go

import (
	"io"
	"os"
	"path/filepath"
)

// Save streams the response to a file, observes LimitBody, and closes the response.
// It writes a temporary file (mode 0600) in the destination directory and renames
// it only after transfer and file/response closing succeed. On failure it attempts to remove
// the temporary file; the destination is not truncated. Rename/replacement semantics
// follow os.Rename on the host platform. Directories must already exist. Save does
// not preserve an existing file's permissions and does not fsync for crash durability.
func (r *Result) Save(path string) error {
	var temp string
	err := r.consume(func(resp *Response) error {
		file, err := os.CreateTemp(filepath.Dir(path), ".requests4go-*")
		if err != nil {
			return err
		}
		temp = file.Name()
		copyErr := r.copyBody(file, resp.Body)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
	if temp != "" {
		defer os.Remove(temp)
	}
	if err != nil {
		return err
	}
	r.err = os.Rename(temp, path)
	return r.err
}

func (r *Result) copyBody(dst io.Writer, src io.Reader) error {
	if !r.limited {
		_, err := io.Copy(dst, src)
		return err
	}
	n, err := io.Copy(dst, io.LimitReader(src, r.limit))
	if err != nil || n < r.limit {
		return err
	}
	return checkBodyOverflow(src)
}

func checkBodyOverflow(src io.Reader) error {
	var extra [1]byte
	n, err := io.ReadFull(src, extra[:])
	if n != 0 {
		return ErrResponseTooLarge
	}
	if err != io.EOF {
		return err
	}
	return nil
}
