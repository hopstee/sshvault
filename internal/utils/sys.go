package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func CheckBinary(name string) error {
	_, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf(
			"Utility %q not found in PATH. You need to install it.",
			name,
		)
	}
	return nil
}

func HostKnown(host string) bool {
	cmd := exec.Command("ssh-keygen", "-F", host)
	return cmd.Run() == nil
}

func GetHostKey(port int, host string) (string, error) {
	cmd := exec.Command("ssh-keyscan", "-p", strconv.Itoa(port), host)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func GetFingerprint(host string) (string, error) {
	cmd := exec.Command("ssh-keyscan", "-p", "22", host)

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	fpCmd := exec.Command("ssh-keygen", "-lf", "-")
	fpCmd.Stdin = bytes.NewReader(out)

	fpOut, err := fpCmd.Output()
	if err != nil {
		return "", err
	}

	return string(fpOut), nil
}

func AddToKnownHosts(data string) error {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".ssh", "known_hosts")

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(data)
	return err
}
