package commands

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"fs-encrypt/internal/archive"
	"fs-encrypt/internal/crypto"
)

var decryptCmd = &cobra.Command{
	Use:   "decrypt [encrypted_file] [destination_dir]",
	Short: "Decrypt a file to a directory",
	Long:  `Decrypts an encrypted file using AES-256-GCM, decompresses with Zstd, and extracts to the destination directory.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		src := args[0]
		dst := args[1]

		// Check if source exists
		info, err := os.Stat(src)
		if os.IsNotExist(err) {
			return fmt.Errorf("source file does not exist: %s", src)
		}
		if info.IsDir() {
			return fmt.Errorf("source must be a file: %s", src)
		}

		// Check if destination exists or create it
		if _, err := os.Stat(dst); os.IsNotExist(err) {
			if err := os.MkdirAll(dst, 0755); err != nil {
				return fmt.Errorf("failed to create destination directory: %w", err)
			}
		}

		// Get password securely
		password, err := readPassword("Enter password: ")
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}

		if len(password) == 0 {
			return fmt.Errorf("password cannot be empty")
		}

		// Open input file
		f, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("failed to open source file: %w", err)
		}
		defer f.Close()

		fmt.Printf("Decrypting %s to %s...\n", src, dst)
		
		// Setup progress bar based on input file size
		bar := progressbar.DefaultBytes(
			info.Size(),
			"Decrypting",
		)
		
		// Wrap reader with progress bar
		// progressbar.NewReader wraps an io.Reader
		// Wait, `progressbar.DefaultBytes` returns a bar.
		// `progressbar.NewReader(r, bar)`? No.
		// `io.TeeReader(r, bar)` works if bar is writer.
		// `progressbar` implements `io.Writer`.
		// So `TeeReader(f, bar)` reads from f and writes to bar.
		barReader := io.TeeReader(f, bar)

		// Create decryption reader
		decReader, err := crypto.NewDecryptReader(barReader, password)
		if err != nil {
			return fmt.Errorf("failed to create decryption reader: %w", err)
		}

		// Create decompression reader
		zstdReader, err := zstd.NewReader(decReader)
		if err != nil {
			return fmt.Errorf("failed to create decompression reader: %w", err)
		}
		defer zstdReader.Close()

		start := time.Now()
		if err := archive.Unpack(zstdReader, dst); err != nil {
			return fmt.Errorf("unpacking failed: %w", err)
		}

		fmt.Printf("\nDone in %v\n", time.Since(start))
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(decryptCmd)
}
