package api

import (
	"github.com/virtual-vgo/vvgo/pkg/clix"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"net/http"
	"net/http/pprof"
)

var logger = log.Logger()

var PublicFiles = "public"

type ServerConfig struct {
	ListenAddress    string
	MaxContentLength int64
	BasicAuthUser    string
	BasicAuthPass    string
}

type Database struct {
	sheets.Sheets
	clix.Clix
}

func NewDatabase(client *storage.Client) *Database {
	sheetsBucket := client.NewBucket(SheetsBucketName)
	sheetsLocker := client.NewLocker(SheetsLockerKey)
	clixBucket := client.NewBucket(ClixBucketName)
	clixLocker := client.NewLocker(ClixLockerKey)

	if sheetsBucket == nil || sheetsLocker == nil || clixBucket == nil || clixLocker == nil {
		return nil
	}

	return &Database{
		Sheets: sheets.Sheets{
			Bucket: sheetsBucket,
			Locker: sheetsLocker,
		},
		Clix: clix.Clix{
			Bucket: clixBucket,
			Locker: clixLocker,
		},
	}
}

func (x *Database) Init() {
	x.Sheets.Init()
	x.Clix.Init()
}

func NewServer(config ServerConfig, database *Database) *http.Server {
	auth := make(basicAuth)
	if config.BasicAuthUser != "" {
		auth[config.BasicAuthUser] = config.BasicAuthPass
	}

	mux := http.NewServeMux()

	mux.Handle("/auth",
		auth.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("authenticated"))
		})),
	)

	// debug endpoints from net/http/pprof
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	mux.Handle("/sheets", auth.Authenticate(SheetsHandler{database.Sheets}))
	mux.Handle("/sheets/", http.RedirectHandler("/sheets", http.StatusMovedPermanently))

	mux.Handle("/clix", auth.Authenticate(ClixHandler{database.Clix}))
	mux.Handle("/clix/", http.RedirectHandler("/clix", http.StatusMovedPermanently))

	downloadHandler := DownloadHandler{
		SheetsBucketName: database.Sheets.Bucket.DownloadURL,
		ClixBucketName:   database.Clix.Bucket.DownloadURL,
	}
	mux.Handle("/download", auth.Authenticate(downloadHandler))

	uploadHandler := UploadHandler{database}
	mux.Handle("/upload", auth.Authenticate(uploadHandler))

	mux.Handle("/version", http.HandlerFunc(Version))
	mux.Handle("/", http.FileServer(http.Dir("public")))

	return &http.Server{
		Addr:     config.ListenAddress,
		Handler:  mux,
		ErrorLog: log.StdLogger(),
	}
}
