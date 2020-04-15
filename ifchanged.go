package ifchanged

import (
	"fmt"
	"os"
)

type DB interface {
	Put(key, value []byte) error
	Has(key []byte) bool
	Get(key []byte) ([]byte, error)
	Sync() error
	Close() error
}

// UsingDB based on a key/value containing sha256 checksum in DB database.
// Runs `executeIfChanged` if checksum has changed.
// If `executeIfChanged` returns error, we don't update sha256.
// The function is not retuning error of executeIfChanged() if happened.
func UsingDB(fileName string, db DB, sha256key []byte, executeIfChanged func() error) error {
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
			return fmt.Errorf("DB error: %w", err)
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
func UsingFile(fileName string, sha256file string, executeIfChanged func() error) error {
	return _ifChangedUsingFile(fileName, sha256file, "", executeIfChanged)
}

// IfChangedOrFileMissingUsingFile additionally executes executeIfChanged
// when particular file is missing. Useful when `executeIfChanged` generates that
// exact file and it serves as a sign that operation is necessary to perform
func UsingFileOrMissing(fileName string, sha256file string, checkMissingFileName string, executeIfChanged func() error) error {
	return _ifChangedUsingFile(fileName, sha256file, checkMissingFileName, executeIfChanged)
}

// _ifChangedUsingFile based on a file that contains sha256 checksum.
// Runs `executeIfChanged` if checksum has changed.
// If `executeIfChanged` returns error, we don't update sha256.
// The function is not retuning error of executeIfChanged() if happened.
func _ifChangedUsingFile(fileName string, sha256file string, checkMissingFileName string, executeIfChanged func() error) error {
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s due to: %w", fileName, err)
		} else {
			return fmt.Errorf("file opening error: %s due to: %w", fileName, err)
		}
	} else if fileInfo.IsDir() {
		return fmt.Errorf("file is a folder: %s due to: %w", fileName, err)
	}
	newSha256, err := GetFileSHA256(fileName)
	if err != nil {
		return fmt.Errorf("sha256 not found: %w", err)
	}
	forceGenerate := false
	if checkMissingFileName != "" {
		checkMissingFileInfo, err := os.Stat(checkMissingFileName)
		if err != nil {
			if os.IsNotExist(err) {
				forceGenerate = true
			} else {
				return fmt.Errorf("file opening error: %s due to: %w", checkMissingFileInfo, err)
			}
		} else if checkMissingFileInfo.IsDir() {
			return fmt.Errorf("checkMissingFileName is a folder: %s due to: %w", checkMissingFileName, err)
		}
	}
	savedSha256 := ""
	if !forceGenerate { // No need to read old sha256 if we have to overwrite it anyway
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
	}
	if forceGenerate || savedSha256 == "" || savedSha256 != newSha256 {
		if err := executeIfChanged(); err == nil { // We only want to update the key if execution happened without error
			err = SaveSHA256(newSha256, sha256file)
			if err != nil {
				return fmt.Errorf("error saving sha256 file: %w", err)
			}
		}
	}
	return nil
}
