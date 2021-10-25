package proxy

import (
	"io"
)

const bufferSize = 1024

// pipe forwards a source to a destination stream with a buffer.
// Closes and exits with the stop flag.
func pipe(src io.ReadCloser, dst io.WriteCloser, stop *bool) error {

	buffer := make([]byte, bufferSize)

	for {
		if *stop {
			src.Close()
			return nil
		}

		err := func() error {

			n, err := src.Read(buffer)
			if err != nil {
				return err
			}

			_, err = dst.Write(buffer[:n])
			return err

		}()
		if err != nil {
			*stop = true
			src.Close()
			return err
		}

	}

	return nil

}
