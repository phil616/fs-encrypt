package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	SaltSize   = 16
	KeySize    = 32 // AES-256
	NonceSize  = 12
	ChunkSize  = 64 * 1024 // 64KB chunks
)

// DeriveKey derives a 32-byte key from a password and salt using Argon2id.
func DeriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, KeySize)
}

// NewEncryptWriter returns a writer that encrypts data written to it.
func NewEncryptWriter(w io.Writer, password string) (io.WriteCloser, error) {
	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	// Write salt first
	if _, err := w.Write(salt); err != nil {
		return nil, err
	}

	key := DeriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &encryptWriter{
		w:   w,
		gcm: gcm,
		buf: make([]byte, 0, ChunkSize),
	}, nil
}

type encryptWriter struct {
	w   io.Writer
	gcm cipher.AEAD
	buf []byte
}

func (e *encryptWriter) Write(p []byte) (int, error) {
	total := len(p)
	for len(p) > 0 {
		n := ChunkSize - len(e.buf)
		if n > len(p) {
			n = len(p)
		}
		e.buf = append(e.buf, p[:n]...)
		p = p[n:]
		if len(e.buf) == ChunkSize {
			if err := e.flush(); err != nil {
				return total - len(p) - n, err // Return bytes consumed before error
			}
		}
	}
	return total, nil
}

func (e *encryptWriter) flush() error {
	if len(e.buf) == 0 {
		return nil
	}

	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return err
	}

	ciphertext := e.gcm.Seal(nil, nonce, e.buf, nil)
	
	// Write chunk length (ciphertext length)
	if err := binary.Write(e.w, binary.LittleEndian, uint32(len(ciphertext))); err != nil {
		return err
	}
	
	// Write nonce
	if _, err := e.w.Write(nonce); err != nil {
		return err
	}

	// Write ciphertext
	if _, err := e.w.Write(ciphertext); err != nil {
		return err
	}

	e.buf = e.buf[:0]
	return nil
}

func (e *encryptWriter) Close() error {
	return e.flush()
}

// NewDecryptReader returns a reader that decrypts data read from r.
func NewDecryptReader(r io.Reader, password string) (io.Reader, error) {
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(r, salt); err != nil {
		return nil, err
	}

	key := DeriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &decryptReader{
		r:   r,
		gcm: gcm,
	}, nil
}

type decryptReader struct {
	r   io.Reader
	gcm cipher.AEAD
	buf []byte
	err error
}

func (d *decryptReader) Read(p []byte) (int, error) {
	if len(d.buf) > 0 {
		n := copy(p, d.buf)
		d.buf = d.buf[n:]
		return n, nil
	}

	if d.err != nil {
		return 0, d.err
	}

	// Read next chunk
	var length uint32
	if err := binary.Read(d.r, binary.LittleEndian, &length); err != nil {
		d.err = err
		if err == io.EOF {
			return 0, io.EOF
		}
		return 0, err
	}

	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(d.r, nonce); err != nil {
		d.err = err
		return 0, err
	}

	ciphertext := make([]byte, length)
	if _, err := io.ReadFull(d.r, ciphertext); err != nil {
		d.err = err
		return 0, err
	}

	plaintext, err := d.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		d.err = err
		return 0, err
	}

	d.buf = plaintext
	n := copy(p, d.buf)
	d.buf = d.buf[n:]
	return n, nil
}
