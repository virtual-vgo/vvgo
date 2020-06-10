//go:generate go run github.com/virtual-vgo/vvgo/tools/version

// This tool is used to upload reference materials like sheet music and click tracks.

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

type Flags struct {
	version  bool
	project  string
	endpoint string
	user     string
	pass     string
	save     bool
}

func (x *Flags) Parse() {
	flag.BoolVar(&x.version, "version", false, "print version and quit")
	flag.StringVar(&x.project, "project", "", "project for these uploads (required)")
	flag.StringVar(&x.endpoint, "endpoint", "https://vvgo.org", "vvgo endpoint")
	flag.BoolVar(&x.save, "save", false, "save local copy (debugging)")
	flag.Parse()
}

var (
	red    = color.New(color.FgRed)
	blue   = color.New(color.FgBlue)
	yellow = color.New(color.FgYellow)
	green  = color.New(color.FgGreen)
	bold   = color.New(color.Bold)
)

type CmdClient struct {
	*api.AsyncClient
	flags Flags
}

func printError(err error) { red.Fprintf(os.Stderr, ":: error: %v\n", err) }

func main() {
	if err := func() error {
		var flags Flags
		flags.Parse()

		if flags.version {
			fmt.Println(string(version.JSON()))
			os.Exit(0)
		}

		// Build a new client
		token := readApiKey()
		client := CmdClient{
			AsyncClient: api.NewAsyncClient(api.AsyncClientConfig{
				ClientConfig: api.ClientConfig{
					ServerAddress: flags.endpoint,
					Token:         token,
				},
				MaxParallel: 32,
				QueueLength: 64,
			}),
			flags: flags,
		}

		// Make sure we can authenticate
		if err := client.Authenticate(); err != nil {
			return fmt.Errorf("unable to authenticate client: %v", err)
		}

		// check if the project exists
		_, err := client.GetProject(flags.project)
		switch err {
		case nil:
			break
		case projects.ErrNotFound:
			return fmt.Errorf("unkown project: %s", flags.project)
		default:
			return err
		}

		// Close the client on interrupt
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		go func() {
			var sig os.Signal
			sig = <-sigCh
			red.Printf(":: received: %s\n", sig)
			shutdown(client, sigCh)
		}()

		// Read and upload files
		reader := bufio.NewReader(os.Stdin)
		for _, fileName := range flag.Args() {
			select {
			case result := <-client.Status():
				printResult(result)
			default:
				client.uploadFile(os.Stdout, reader, flags.project, fileName)
			}
		}
		shutdown(client, sigCh)
		return nil
	}(); err != nil {
		red.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func readApiKey() string {
	fmt.Print(":: enter token: ")
	tokenBytes, _ := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	return string(tokenBytes)
}

var shutdownOnce sync.Once

func shutdown(client CmdClient, sigCh chan os.Signal) {
	shutdownOnce.Do(func() {
		yellow.Fprintln(os.Stdout, ":: waiting for uploads to complete... Ctrl+C again to force quit")
		done := make(chan struct{})

		go client.Close()

		go func() {
			for result := range client.Status() {
				printResult(result)
			}
			close(done)
		}()
		client.Backup()

		select {
		case sig := <-sigCh:
			red.Printf(":: received: %s\n", sig)
			os.Exit(1)
		case <-done:
			yellow.Println(":: closed!")
			os.Exit(0)
		}
	})
}

func printResult(result api.UploadStatus) {
	if result.Code != http.StatusOK {
		printError(fmt.Errorf(":: upload failed -- %s received non-200 status: `%d: %s`",
			result.FileName, result.Code, result.Error))
	} else {
		green.Printf(":: upload successful -- %s\n", result.FileName)
	}
}

func (x *CmdClient) uploadFile(writer io.Writer, reader *bufio.Reader, project string, fileName string) {
	for {
		blue.Printf(":: %s %s\n", "found:", bold.Sprint(fileName))

		var upload api.Upload
		err := readUpload(writer, reader, &upload, project, fileName)
		switch {
		case err == ErrSkipped:
			blue.Fprintln(writer, ":: skipping...")
			return
		case err != nil:
			printError(err)
			if yesNo(writer, reader, "try again? (；一ω一||)") {
				continue
			} else {
				return
			}
		}

		// render results
		gotParts := upload.Parts()
		fmt.Fprintf(writer, ":: upload creates %s for:\n", upload.UploadType)
		for _, part := range gotParts {
			p := color.New(color.FgWhite)
			p.Fprintln(writer, " * "+part.String())
		}

		if yesNo(writer, reader, "is this ok") {
			doUpload(x.AsyncClient, upload, x.flags.save)
			return
		}
	}
}

var ErrSkipped = errors.New("skipped")

func readUpload(writer io.Writer, reader *bufio.Reader, upload *api.Upload, project string, fileName string) error {
	fileBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	contentType := readMediaType(fileBytes)
	uploadType, err := readUploadType(writer, reader, contentType)
	if err != nil {
		return err
	}
	fmt.Fprintf(writer, ":: upload type: %s\n", bold.Sprint(uploadType))
	fmt.Fprintf(writer, ":: %s | %s\n",
		color.New(color.Italic).Sprint("leave empty to skip"),
		color.New(color.Italic).Sprint("Ctrl+C to quit"))
	partNames := readPartNames(writer, reader)
	*upload = api.Upload{
		UploadType:  uploadType,
		PartNames:   partNames,
		Project:     project,
		FileName:    fileName,
		FileBytes:   fileBytes,
		ContentType: contentType,
	}

	if partNames == nil {
		return ErrSkipped
	}
	return upload.Validate()
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

func readUploadType(writer io.Writer, reader *bufio.Reader, mediaType string) (api.UploadType, error) {
	switch true {
	case strings.HasPrefix(mediaType, "application/pdf"):
		return api.UploadTypeSheets, nil
	case strings.HasPrefix(mediaType, "audio/"):
		return api.UploadTypeClix, nil
	default:
		return "", fmt.Errorf("i don't know how to handle media type: `%s`. (´･ω･`)", mediaType)
	}
}

func readPartNames(writer io.Writer, reader *bufio.Reader) []string {
	fmt.Fprint(writer, ":: part names (ex. trumpet, flute): ")
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

func doUpload(client *api.AsyncClient, upload api.Upload, save bool) {
	if save {
		allUploads = append(allUploads, upload)
		f, err := os.OpenFile("__uploader.out", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if err == nil {
			json.NewEncoder(f).Encode(&allUploads)
			f.Close()
		}
	}
	client.Upload(upload)
	yellow.Printf(":: upload queued -- %s\n", upload.FileName)
}
