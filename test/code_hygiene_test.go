package test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCodeHygiene(t *testing.T) {
	for _, tt := range []struct {
		cmd   string
		args  []string
		files []string
	}{
		{
			cmd:   "gofmt",
			args:  []string{"-d"},
			files: filesWithExtension("..", ".go"),
		},
		{
			cmd:   "shellcheck",
			files: filesWithExtension("..", ".sh"),
		},
	} {
		t.Run(tt.cmd, func(t *testing.T) {
			if len(tt.files) == 0 {
				t.Log("no files to test!")
				return
			}
			cmd := exec.Command(tt.cmd, append(tt.args, tt.files...)...)
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			cmd.Stdout = &stdout
			if err := cmd.Run(); err != nil {
				t.Errorf("%s failed: %v", tt.cmd, err)
			}
			if stderr.Len() != 0 {
				t.Errorf("%s", stderr.String())
			}
			if stdout.Len() != 0 {
				t.Errorf("%s", stdout.String())
			}
		})
	}
}

func filesWithExtension(root string, ext string) []string {
	var fileNames []string
	err := filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		// return errors immediately
		if err != nil {
			panic(fmt.Sprintf("filepath.Walk failed: %v", err))
		}

		// check for files matching the extension
		if filepath.Ext(f.Name()) == ext {
			fileNames = append(fileNames, path)
		}
		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("filepath.Walk failed: %v", err))
	}
	return fileNames
}
