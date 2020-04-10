//go:generate go run github.com/virtual-vgo/vvgo/tools/version

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/sheet"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Flags struct {
	version    bool
	project    string
	uploadType string
	endpoint   string
	user       string
	pass       string
}

func (x *Flags) Parse() {
	flag.BoolVar(&x.version, "version", false, "print version and quit")
	flag.StringVar(&x.project, "project", "", "project for these uploads (required)")
	flag.StringVar(&x.uploadType, "upload-type", "", "type of upload: sheets, clix")
	flag.StringVar(&x.endpoint, "endpoint", "https://vvgo.org/upload", "upload endpoint")
	flag.StringVar(&x.user, "user", "admin", "basic auth username")
	flag.StringVar(&x.pass, "pass", "admin", "basic auth password")
	flag.Parse()
}

var (
	red    = color.New(color.FgRed)
	blue   = color.New(color.FgBlue)
	yellow = color.New(color.FgYellow)
	green  = color.New(color.FgGreen)
)

func main() {
	if err := func() error {
		var flags Flags
		flags.Parse()

		if flags.version {
			fmt.Println(version.String())
			os.Exit(0)
		}

		if flags.project != "01-snake-eater" {
			return fmt.Errorf("unkown project: %s", flags.project)
		}

		reader := bufio.NewReader(os.Stdin)

		// loop over each file name
		for _, fileName := range flag.Args() {
			uploadSheet(reader, flags, fileName)
		}

		return nil
	}(); err != nil {
		red.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func uploadSheet(reader *bufio.Reader, flags Flags, fileName string) {
	blue.Printf(":: found `%s`\n", fileName)

	if ! yesNo(reader, "upload this file") {
		blue.Println("skipping...")
		return
	}

	// read the file
	fileBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		printError(err)
		return
	}

	var uploads []api.Upload
	for {
		// read the part numbers
		fmt.Printf(":: please enter part numbers (ex 1, 2): ")
		rawNumbers, _ := reader.ReadString('\n')
		var partNumbers []int
		for _, raw := range strings.Split(rawNumbers, ",") {
			real, err := strconv.Atoi(strings.TrimSpace(raw))
			if err != nil {
				printError(err)
			} else {
				partNumbers = append(partNumbers, real)
			}
		}

		// read the part names
		fmt.Printf(":: please enter part names (ex trumpet, flute): ")
		rawNames, _ := reader.ReadString('\n')
		partNames := strings.Split(rawNames, ",")
		for i := range partNames {
			partNames[i] = strings.ToLower(strings.TrimSpace(partNames[i]))
		}

		// make the upload request
		upload := api.Upload{
			UploadType: api.UploadTypeSheets,
			ClixUpload: nil,
			SheetsUpload: &api.SheetsUpload{
				PartNames:   partNames,
				PartNumbers: partNumbers,
			},
			Project:     flags.project,
			FileName:    fileName,
			FileBytes:   fileBytes,
			ContentType: "application/pdf",
		}

		// validate the sheets locally
		var gotSheets []sheet.Sheet
		for _, sheet := range upload.ToSheets() {
			if err := sheet.Validate(); err != nil {
				printError(err)
			} else {
				gotSheets = append(gotSheets, sheet)
			}
		}

		// render what the results will look like
		fmt.Println(":: this will create the following entries:")
		for _, sheet := range gotSheets {
			fmt.Println(sheet.String())
		}
		if yesNo(reader, "is this ok") {
			uploads = append(uploads, upload)
			doUpload(flags, uploads)
			return
		}
	}
}

func yesNo(reader *bufio.Reader, pre string) bool {
	yellow.Printf(":: %s [Y/n]? ", pre)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "" || answer == "y" || answer == "yes"
}

func doUpload(flags Flags, uploads []api.Upload) {
	if len(uploads) == 0 {
		return
	}

	var buffer bytes.Buffer
	json.NewEncoder(&buffer).Encode(&uploads)
	req, err := http.NewRequest(http.MethodPost, flags.endpoint, &buffer)
	if err != nil {
		printError(err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(flags.user, flags.pass)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		printError(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		printError(fmt.Errorf("http request received non-200 status: `%d: %s`", resp.StatusCode, bytes.TrimSpace(body)))
		return
	}

	var results []api.UploadStatus
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		printError(err)
	}
	for _, result := range results {
		if result.Code != http.StatusOK {
			printError(fmt.Errorf("file %s received non-200 status: `%d: %s`", result.FileName, result.Code, result.Error))
		} else {
			green.Printf(":: file %s uploaded successfully!\n", result.FileName)
		}
	}
}

func printError(err error) {
	red.Fprintf(os.Stderr, ":: error: %v\n", err)
}
