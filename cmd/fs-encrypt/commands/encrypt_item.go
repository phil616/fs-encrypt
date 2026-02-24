package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var encryptItemCmd = &cobra.Command{
	Use:   "encrypt-item [source_file] [output_file]",
	Short: "Encrypt a single file",
	Long:  `Encrypts a single file using the same format as directory encryption.`,
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

		return performEncryption(src, dst, password)
	},
}

func init() {
	rootCmd.AddCommand(encryptItemCmd)
}
