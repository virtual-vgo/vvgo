//go:generate go run github.com/virtual-vgo/vvgo/tools/version

// This tool is used to upload reference materials like sheet music and click tracks.

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/projects"
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
	flag.StringVar(&x.user, "user", "vvgo-dev", "basic auth username")
	flag.StringVar(&x.pass, "pass", "vvgo-dev", "basic auth password")
	flag.Parse()
}

var (
	red    = color.New(color.FgRed)
	blue   = color.New(color.FgBlue)
	yellow = color.New(color.FgYellow)
	green  = color.New(color.FgGreen)
	bold   = color.New(color.Bold)
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
			uploadFile(client, os.Stdout, reader, flags.project, fileName)
		}

		f, err := os.OpenFile("__uploader.out", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		json.NewEncoder(f).Encode(&allUploads)
		f.Close()
		return nil
	}(); err != nil {
		red.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func uploadFile(client *api.Client, writer io.Writer, reader *bufio.Reader, project string, fileName string) {
	for {
		blue.Printf(":: found `%s`\n", fileName)

		var upload api.Upload
		if ok := readUpload(writer, reader, &upload, project, fileName); !ok {
			return
		}

		if err := upload.Validate(); err != nil {
			printError(err)
			if yesNo(writer, reader, "try again? (；一ω一||)") {
				continue
			} else {
				return
			}
		}

		// render results
		gotParts := upload.Parts()
		fmt.Fprintf(writer, ":: upload creates %s for the following parts:\n", upload.UploadType)
		for _, part := range gotParts {
			fmt.Fprintln(writer, " * "+part.String())
		}
		if yesNo(writer, reader, "is this ok") {
			doUpload(client, upload)
			return
		}
	}
}

func readUpload(writer io.Writer, reader *bufio.Reader, dest *api.Upload, project string, fileName string) bool {
	fileBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		printError(err)
		return false
	}
	contentType := readMediaType(fileBytes)
	uploadType := readUploadType(writer, reader, contentType)
	if uploadType == "" {
		return false
	}
	fmt.Fprintf(writer, ":: upload type: %s | %s\n",
		bold.Sprint(uploadType),
		color.New(color.Italic).Sprint("leave part names or numbers empty to skip"))
	partNames := readPartNames(writer, reader)
	partNumbers := readPartNumbers(writer, reader)
	if partNames == nil || partNumbers == nil {
		blue.Fprintln(writer, ":: skipping...")
		return false
	}
	*dest = api.Upload{
		UploadType:  uploadType,
		PartNames:   partNames,
		PartNumbers: partNumbers,
		Project:     project,
		FileName:    fileName,
		FileBytes:   fileBytes,
		ContentType: contentType,
	}
	return true
}

func yesNo(writer io.Writer, reader *bufio.Reader, pre string) bool {
	yellow.Fprintf(writer, ":: %s [Y/n]? ", pre)
	answer, err := reader.ReadString('\n')
	if err != nil {
		printError(err)
		return false
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "" || answer == "y" || answer == "yes"
}

func readMediaType(fileBytes []byte) string {
	// guess upload type based on the file contents
	return http.DetectContentType(fileBytes)
}

func readUploadType(writer io.Writer, reader *bufio.Reader, mediaType string) api.UploadType {
	switch true {
	case strings.HasPrefix(mediaType, "application/pdf"):
		return api.UploadTypeSheets
	case strings.HasPrefix(mediaType, "audio/"):
		return api.UploadTypeClix
	default:
		printError(fmt.Errorf("i don't know how to handle media type: `%s`. (´･ω･`)", mediaType))
	}
	return ""
}

func readPartNumbers(writer io.Writer, reader *bufio.Reader) []uint8 {
	fmt.Fprintf(writer, ":: part numbers (ex. 1, 2): ")
	rawNumbers, err := reader.ReadString('\n')
	if err != nil {
		printError(err)
		return nil
	}

	rawNumbers = strings.TrimSpace(rawNumbers)
	if rawNumbers == "" {
		return nil
	}

	var partNums []uint8
	for _, raw := range strings.Split(rawNumbers, ",") {
		bigNum, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 8)
		num := uint8(bigNum)
		switch {
		case err != nil:
			printError(err)
			return nil
		default:
			partNums = append(partNums, num)
		}
	}
	return partNums
}

func readPartNames(writer io.Writer, reader *bufio.Reader) []string {
	fmt.Fprintf(writer, ":: part names (ex. trumpet, flute): ")
	rawNames, err := reader.ReadString('\n')
	if err != nil {
		printError(err)
		return nil
	}

	rawNames = strings.TrimSpace(rawNames)
	if rawNames == "" {
		return nil
	}

	var names []string
	for _, name := range strings.Split(rawNames, ",") {
		name := strings.ToLower(strings.TrimSpace(name))
		names = append(names, name)
	}
	return names
}

var allUploads []api.Upload

func doUpload(client *api.Client, upload api.Upload) {
	allUploads = append(allUploads, upload)

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
