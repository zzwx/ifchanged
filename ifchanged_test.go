package ifchanged

import (
	"testing"
)

func TestIfChangedOrFileMissingUsingFile(t *testing.T) {
	_, err := ReadFileAsString("./examples/example1.generated")
	if err == nil {
		t.Errorf("./examples/example1.generated shouldn't exist before test begins")
	}

	visited := false
	if err := NewIf().
		Changed("./examples/example1.go", "./examples/example1.go.sha256").
		Missing("./examples/example1.generated").Execute(
		func() error {
			visited = true
			return nil
		}); err != nil {
		t.Errorf("IfChangedOrFileMissingUsingFile() error = %v", err)
	}
	if !visited {
		t.Errorf("IfChangedOrFileMissingUsingFile() error: function wasn't visited")
	}
}
