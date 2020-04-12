//go:generate go run github.com/virtual-vgo/vvgo/tools/version

// This tool is used to upload reference materials like sheet music and click tracks.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/clix"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Flags struct {
	version  bool
	project  string
	endpoint string
	user     string
	pass     string
}

func (x *Flags) Parse() {
	flag.BoolVar(&x.version, "version", false, "print version and quit")
	flag.StringVar(&x.project, "project", "", "project for these uploads (required)")
	flag.StringVar(&x.endpoint, "endpoint", "https://vvgo.org", "vvgo endpoint")
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

func printError(err error) { red.Fprintf(os.Stderr, ":: error: %v\n", err) }

func main() {
	if err := func() error {
		var flags Flags
		flags.Parse()

		if flags.version {
			fmt.Println(version.String())
			os.Exit(0)
		}

		if projects.Exists(flags.project) == false {
			return fmt.Errorf("unkown project: %s", flags.project)
		}

		client := api.NewClient(api.ClientConfig{
			ServerAddress: flags.endpoint,
			BasicAuthUser: flags.user,
			BasicAuthPass: flags.pass,
		})

		if err := client.Authenticate(); err != nil {
			return fmt.Errorf("unable to authenticate client: %v", err)
		}

		reader := bufio.NewReader(os.Stdin)
		for _, fileName := range flag.Args() {
			uploadFile(client, reader, flags.project, fileName)
		}
		return nil
	}(); err != nil {
		red.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func uploadFile(client *api.Client, reader *bufio.Reader, project string, fileName string) {
	for {
		blue.Printf(":: found `%s`\n", fileName)

		if !yesNo(os.Stdout, reader, "upload this file") {
			blue.Println("skipping...")
			return
		}

		// read the file
		fileBytes, err := ioutil.ReadFile(fileName)
		if err != nil {
			printError(err)
			return
		}

		// guess upload type based on the file contents
		contentType := http.DetectContentType(fileBytes)

		// start the upload request
		upload := api.Upload{
			Project:     project,
			FileName:    fileName,
			FileBytes:   fileBytes,
			ContentType: contentType,
		}

		switch true {
		case sheets.ValidMediaType(contentType):
			if !yesNo(os.Stdout, reader, "this is a music sheet") {
				red.Println(":: i don't know what this is. つ´Д`)つ")
				return
			}
			readSheetUpload(os.Stdout, reader, &upload)
		case clix.ValidMediaType(contentType):
			if !yesNo(os.Stdout, reader, "this is a click track") {
				red.Println(":: i don't know what this is. (;´д｀)")
				return
			}
			readClickUpload(os.Stdout, reader, &upload)
		default:
			red.Printf(":: i don't know how to handle media type: `%s`. (´･ω･`)", contentType)
			return
		}

		if yesNo(os.Stdout, reader, "is this ok") {
			doUpload(client, upload)
			return
		}
	}
}

func readClickUpload(writer io.Writer, reader *bufio.Reader, dest *api.Upload) {
	partNumbers := readPartNumbers(writer, reader)
	partNames := readPartNames(writer, reader)

	dest.UploadType = api.UploadTypeClix
	dest.ClixUpload = &api.ClixUpload{
		PartNames:   partNames,
		PartNumbers: partNumbers,
	}

	// validate the sheets locally
	var gotClix []clix.Click
	for _, click := range dest.Clix() {
		if err := click.Validate(); err != nil {
			printError(err)
		} else {
			gotClix = append(gotClix, click)
		}
	}

	// render results
	fmt.Fprintln(writer, ":: this will create the following clix:")
	for _, click := range gotClix {
		fmt.Fprintln(writer, click.String())
	}
}

func readSheetUpload(writer io.Writer, reader *bufio.Reader, dest *api.Upload) {
	partNumbers := readPartNumbers(writer, reader)
	partNames := readPartNames(writer, reader)

	dest.UploadType = api.UploadTypeSheets
	dest.SheetsUpload = &api.SheetsUpload{
		PartNames:   partNames,
		PartNumbers: partNumbers,
	}

	// validate the sheets locally
	var gotSheets []sheets.Sheet
	for _, sheet := range dest.Sheets() {
		if err := sheet.Validate(); err != nil {
			printError(err)
		} else {
			gotSheets = append(gotSheets, sheet)
		}
	}

	// render results
	fmt.Fprintln(writer, ":: this will create the following sheets:")
	for _, sheet := range gotSheets {
		fmt.Fprintln(writer, sheet.String())
	}
}

func yesNo(writer io.Writer, reader *bufio.Reader, pre string) bool {
	yellow.Fprintf(writer, ":: %s [Y/n]? ", pre)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "" || answer == "y" || answer == "yes"
}

func readPartNumbers(writer io.Writer, reader *bufio.Reader) []uint8 {
	fmt.Fprintf(writer, ":: please enter part numbers (ex 1, 2): ")
	rawNumbers, _ := reader.ReadString('\n')
	var partNumbers []uint8
	for _, raw := range strings.Split(rawNumbers, ",") {
		number, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 8)
		if err != nil {
			printError(err)
		} else {
			partNumbers = append(partNumbers, uint8(number))
		}
	}
	return partNumbers
}

func readPartNames(writer io.Writer, reader *bufio.Reader) []string {
	fmt.Fprintf(writer, ":: please enter part names (ex trumpet, flute): ")
	rawNames, _ := reader.ReadString('\n')
	partNames := strings.Split(rawNames, ",")
	for i := range partNames {
		partNames[i] = strings.ToLower(strings.TrimSpace(partNames[i]))
	}
	return partNames
}

func doUpload(client *api.Client, upload api.Upload) {
	results, err := client.Upload(upload)
	if err != nil {
		printError(err)
		return
	}
	for _, result := range results {
		if result.Code != http.StatusOK {
			printError(fmt.Errorf("file %s received non-200 status: `%d: %s`", result.FileName, result.Code, result.Error))
		} else {
			green.Printf(":: file %s uploaded successfully!\n", result.FileName)
		}
	}
}
