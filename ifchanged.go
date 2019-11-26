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

	"github.com/prologic/bitcask"
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

// IfChangedUsingBitcask based on a key/value containing sha256 checksum in bitcask database.
// Runs `executeIfChanged` if checksum has changed.
// If `executeIfChanged` returns error, we don't update sha256.
// The function is not retuning error of executeIfChanged() if happened.
func IfChangedUsingBitcask(fileName string, db *bitcask.Bitcask, sha256key []byte, executeIfChanged func() error) error {
	fileInfo, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s due to: %w", fileName, err)
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("file is a folder: %s due to: %w", fileName, err)
	}
	savedSha256 := ""
	if db.Has(sha256key) {
		val, err := db.Get(sha256key)
		if err != nil {
			return fmt.Errorf("bitcask error: %w", err)
		}
		if val == nil {
		} else {
			savedSha256 = string(val)
		}
	}

	newSha256, err := GetFileSHA256(fileName)
	if err != nil {
		return fmt.Errorf("sha256 not found: %w", err)
	}
	if savedSha256 == "" || savedSha256 != newSha256 {
		if err := executeIfChanged(); err == nil { // We only want to update the key if execution happened without error
			err = db.Put(sha256key, []byte(newSha256))
			if err != nil {
				return fmt.Errorf("error saving sha256 db value: %w", err)
			}
			err = db.Sync()
			if err != nil {
				return fmt.Errorf("error syncing sha256 db value: %w", err)
			}
		}
	}
	return nil
}

// IfChangedUsingFile based on a file that contains sha256 checksum.
// Runs `executeIfChanged` if checksum has changed.
// If `executeIfChanged` returns error, we don't update sha256.
// The function is not retuning error of executeIfChanged() if happened.
func IfChangedUsingFile(fileName string, sha256file string, executeIfChanged func() error) error {
	fileInfo, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s due to: %w", fileName, err)
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("file is a folder: %s due to: %w", fileName, err)
	}
	savedSha256 := ""
	sha256FileInfo, err := os.Stat(sha256file)
	if err == nil && sha256FileInfo.IsDir() {
		return fmt.Errorf("sha256 is a folder: %w", err)
	}
	if !os.IsNotExist(err) && !sha256FileInfo.IsDir() {
		savedSha256, err = ReadFileAsString(sha256file)
		if err != nil {
			return fmt.Errorf("error reading existing sha256 file: %w", err)
		}
	}
	newSha256, err := GetFileSHA256(fileName)
	if err != nil {
		return fmt.Errorf("sha256 not found: %w", err)
	}
	if savedSha256 == "" || savedSha256 != newSha256 {
		if err := executeIfChanged(); err == nil { // We only want to update the key if execution happened without error
			err = SaveSHA256(newSha256, sha256file)
			if err != nil {
				return fmt.Errorf("error saving sha256 file: %w", err)
			}
		}
	}
	return nil
}
