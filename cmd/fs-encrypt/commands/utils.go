package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	
	// If stdin is a terminal, use secure read
	if term.IsTerminal(int(syscall.Stdin)) {
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println() // Print newline after input
		if err != nil {
			return "", err
		}
		return string(bytePassword), nil
	}

	// Fallback for non-terminal input (e.g. pipe)
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(password), nil
}
