package ifchanged

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

func GetFileSHA256(fileName string) (string, error) {
	hasher := sha256.New()
	s, err := ioutil.ReadFile(fileName)
	hasher.Write(s)
	if err != nil {
		return "", fmt.Errorf("error finding sha256: %w", err)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func ReadFileAsString(fileName string) (string, error) {
	s, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}
	return string(s), nil
}

func SaveSHA256(sha256 string, fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("create sha256 error: %w", err)
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	_, err = fmt.Fprintf(w, `%s`, sha256)
	if err != nil {
		return fmt.Errorf("save sha256 error: %w", err)
	}
	err = w.Flush()
	if err != nil {
		return fmt.Errorf("save sha256 error: %w", err)
	}
	return nil
}

// ExecuteCommand is a helper function to run `exe` file with the `arg`s and
// in case of error, returns a fully formatted error containing both `stderr` and `stdout`
func ExecuteCommand(exe string, arg ...string) error {
	cmd := exec.Command(exe, arg...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run() // Block the thread until it finishes
	if err != nil {
		return fmt.Errorf("execute command error: stderr: \"%s\", stdout: \"%s\" due to %w", stderr.String(), stdout.String(), err)
	}
	return nil
	// cmd.Wait() // Wait is only needed if we cmd.Start()
}
