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

// If is designed for human-readable chaining of "ifs" of changed or missing files cases in which Execute should be called.
// If can be either initialized with NewIf() or as (&ifchanged.If{}) for immediate chaining.
type If struct {
	pairs   []FileSHA256Pair
	missing []string
}

// NewIf initializes new If structure. It is easier to read when used in chaining.
// If can be used as-is, but it has to be wrapped into (&ifchanged.If{}) in order to become immediately chainable.
func NewIf() *If {
	return &If{}
}

// Changed appends a fileName / sha256 file rule to If structure to check against for modified files,
// in which case a function provided with Execute will be called
func (i *If) Changed(fileName, sha256file string) *If {
	i.pairs = append(i.pairs, FileSHA256Pair{
		FileName:   fileName,
		Sha256file: sha256file,
	})
	return i
}

// Missing appends fileName(s) that serve as checks for missing files,
// in which case a function provided with Execute will be called.
func (i *If) Missing(fileName ...string) *If {
	i.missing = append(i.missing, fileName...)
	return i
}

// Execute is expected to be a final call in the chain with Changed / Missing as it doesn't return If back.
// Error returned never comes from calling f. If f does return error, then none of the sha256 files, provided using Changed,
// will be updated.
// Execute doesn't invalidate the If struct, so additional Changed / Missing / Execute are possible on top of existing data.
func (i *If) Execute(f func() error) error {
	return usingFilesOrMissing(i.pairs, i.missing, f)
}

// UsingDB based on a key/value containing sha256 checksum in DB database.
// Runs execute if checksum has changed.
// When execute returns error, we don't update sha256.
// The function does not return any error out of provided execute function.
func UsingDB(fileName string, db DB, sha256key []byte, execute func() error) error {
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
		if err := execute(); err == nil { // We only want to update the key if execution happened without error
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

// UsingFile checks for a sha256 checksum in a sha256file.
// Runs executeIfChanged if new checksum doesn't match previous sha256.
// executeIfChanged can return an error, in which case the sha256file generation / update will be skipped.
// The function does not return any error from executeIfChanged. It has to be handled by the caller.
func UsingFile(fileName string, sha256file string, executeIfChanged func() error) error {
	return usingFilesOrMissing([]FileSHA256Pair{{
		FileName:   fileName,
		Sha256file: sha256file,
	}}, nil, executeIfChanged)
}

// Deprecated: Use UsingFiles or If struct, that work for both a list of modified and a list of possibly missing files.
//
// UsingFileOrMissing acts as UsingFile, but additionally executes executeIfChanged
// when a file checkMissingFileName is missing. This assumes that executeIfChanged generates the
// file, so it serves as a sign that operation is necessary to perform.
func UsingFileOrMissing(fileName string, sha256file string, checkMissingFileName string, execute func() error) error {
	return usingFilesOrMissing([]FileSHA256Pair{{
		FileName:   fileName,
		Sha256file: sha256file,
	}}, []string{checkMissingFileName}, execute)
}

// UsingFilesOrMissingFiles acts as UsingFile, but additionally executes executeIfChanged
// when any of the files listed with checkMissingFileName is missing. This assumes that executeIfChanged generates the
// file(s), so it serves as a sign that operation is necessary to perform. UsingFileOrMissing does that for a just one file (kept for backward compatibility).
func UsingFiles(fileSHA256Pairs []FileSHA256Pair, checkMissingFiles []string, execute func() error) error {
	return usingFilesOrMissing(fileSHA256Pairs, checkMissingFiles, execute)
}

type FileSHA256Pair struct {
	FileName      string
	Sha256file    string
	currentSha256 string
	newSha256     string
}

// usingFilesOrMissing based on a file that contains sha256 checksum.
// Runs `executeIfChanged` if checksum has changed.
// If `executeIfChanged` returns error, we don't update sha256.
// The function is not retuning error of executeIfChanged() if happened.
func usingFilesOrMissing(fileSHA256Pairs []FileSHA256Pair, missingFiles []string, executeIfChanged func() error) error {
	missingDetected := false
	for _, f := range missingFiles {
		checkMissingFileInfo, err := os.Stat(f)
		if err != nil {
			if os.IsNotExist(err) {
				missingDetected = true
				break // One file missing is enough to force regeneration
			} else {
				return fmt.Errorf("file opening error: %s due to: %w", checkMissingFileInfo, err)
			}
		} else if checkMissingFileInfo.IsDir() {
			return fmt.Errorf("checkMissingFileName is a folder: %s due to: %w", f, err)
		}
	}

	changeDetected := false

	for i, filePair := range fileSHA256Pairs {
		fileName := filePair.FileName
		sha256file := filePair.Sha256file
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
		fileSHA256Pairs[i].newSha256 = newSha256
		currentSha256 := ""
		if !missingDetected { // No need to read old sha256 if we have to overwrite it anyway
			sha256FileInfo, err := os.Stat(sha256file)
			if err == nil && sha256FileInfo.IsDir() {
				return fmt.Errorf("sha256 is a folder: %w", err)
			}
			if !os.IsNotExist(err) && !sha256FileInfo.IsDir() {
				currentSha256, err = ReadFileAsString(sha256file)
				if err != nil {
					return fmt.Errorf("error reading existing sha256 file: %w", err)
				}
				fileSHA256Pairs[i].currentSha256 = currentSha256
			}
		}
		if currentSha256 == "" || currentSha256 != newSha256 {
			changeDetected = true
		}
	}
	if missingDetected || changeDetected {
		if err := executeIfChanged(); err == nil { // We only want to update the key if execution happened without error
			for _, f := range fileSHA256Pairs {
				err = SaveSHA256(f.newSha256, f.Sha256file)
				if err != nil {
					return fmt.Errorf("error saving sha256 file: %w", err)
				}
			}
		}
	}
	return nil
}
