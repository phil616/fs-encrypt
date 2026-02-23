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

var encryptCmd = &cobra.Command{
	Use:   "encrypt [source_dir] [output_file]",
	Short: "Encrypt a directory",
	Long:  `Encrypts a directory recursively with Zstd compression (max level) and AES-256-GCM encryption.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		src := args[0]
		dst := args[1]

		// Check if source exists
		info, err := os.Stat(src)
		if os.IsNotExist(err) {
			return fmt.Errorf("source directory does not exist: %s", src)
		}
		if !info.IsDir() {
			return fmt.Errorf("source must be a directory: %s", src)
		}

		// Get password securely
		password, err := readPassword("Enter password: ")
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}

		confirm, err := readPassword("Confirm password: ")
		if err != nil {
			return fmt.Errorf("failed to read password confirmation: %w", err)
		}

		if password != confirm {
			return fmt.Errorf("passwords do not match")
		}

		if len(password) == 0 {
			return fmt.Errorf("password cannot be empty")
		}

		// Create output file
		f, err := os.Create(dst)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()

		fmt.Printf("Encrypting %s to %s...\n", src, dst)

		// Setup progress bar
		// Use a spinner based on output bytes.
		bar := progressbar.NewOptions64(
			-1,
			progressbar.OptionSetDescription("Encrypting (bytes written)"),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(15),
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionSpinnerType(14),
		)

		// Chain: File+Bar <- Encrypt <- Zstd <- Tar <- Dir

		// 1. Wrap file with progress bar writer
		// io.MultiWriter writes to both f and bar.
		// Note: bar implements io.Writer, so it counts bytes written to it.
		barWriter := io.MultiWriter(f, bar)

		// 2. Encrypt Writer
		encWriter, err := crypto.NewEncryptWriter(barWriter, password)
		if err != nil {
			return fmt.Errorf("failed to create encryption writer: %w", err)
		}
		// We must close encWriter to flush the last chunk
		defer encWriter.Close()

		// 3. Compression Writer
		// SpeedBestCompression provides highest compression ratio
		zstdWriter, err := zstd.NewWriter(encWriter, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
		if err != nil {
			return fmt.Errorf("failed to create compression writer: %w", err)
		}
		defer zstdWriter.Close()

		start := time.Now()
		// 4. Pack (Tar)
		if err := archive.Pack(src, zstdWriter); err != nil {
			return fmt.Errorf("packing failed: %w", err)
		}

		// Explicit close to catch errors and flush data
		if err := zstdWriter.Close(); err != nil {
			return fmt.Errorf("compression close failed: %w", err)
		}
		if err := encWriter.Close(); err != nil {
			return fmt.Errorf("encryption close failed: %w", err)
		}
		
		bar.Finish()
		fmt.Printf("\nDone in %v\n", time.Since(start))
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(encryptCmd)
}
