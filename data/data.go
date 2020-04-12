//go:generate go run github.com/rakyll/statik -src . -f -include *.txt

package data

// Package for access static data
// Run `go generate` to package up the data files
// Add a getter method to export your file data

import (
	"github.com/rakyll/statik/fs"
	_ "github.com/virtual-vgo/vvgo/data/statik"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"io/ioutil"
	"net/http"
	"strings"
)

var logger = log.Logger()

var validPartNames map[string]struct{}

func ValidPartNames() map[string]struct{} { return validPartNames }

func init() {
	statikFS, err := fs.New()
	if err != nil {
		logger.Fatal(err)
	}
	validPartNames = readValidPartNames(statikFS)
}

func readValidPartNames(fs http.FileSystem) map[string]struct{} {
	fileBytes := mustRead(fs, "/valid_part_names.txt")
	rawNames := strings.Split(string(fileBytes), "\n")
	names := make(map[string]struct{}, len(rawNames))
	for _, raw := range rawNames {
		raw = strings.TrimSpace(raw)
		if strings.HasPrefix(raw, "#") {
			continue
		}
		names[strings.ToLower(raw)] = struct{}{}
	}
	return names
}

func mustRead(fs http.FileSystem, name string) []byte {
	// read valid part names
	r, err := fs.Open(name)
	if err != nil {
		logger.Fatal(err)
	}
	defer r.Close()

	// read con
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		logger.Fatal(err)
	}
	return contents
}
