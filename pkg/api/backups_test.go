package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"testing"
)

func TestBackupHandler_ServeHTTP(t *testing.T) {
	t.Run("view", func(t *testing.T) {
		handler := BackupHandler{
			Database: &Database{Parts: newParts()},
			Backups:  newBucket(t),
		}

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(recorder, request)
		got := recorder.Result()
		assert.Equal(t, http.StatusOK, got.StatusCode)
		assertEqualHTML(t, mustReadFile(t, "testdata/backups.html"), recorder.Body.String())
	})

	t.Run("backup", func(t *testing.T) {
		ctx := context.Background()
		handler := BackupHandler{
			Database: &Database{Parts: newParts()},
			Backups:  newBucket(t),
		}

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/", nil)
		request.Form = make(url.Values)
		request.Form.Add("cmd", "backup")
		handler.ServeHTTP(recorder, request)

		got := recorder.Result()
		assert.Equal(t, http.StatusOK, got.StatusCode)
		assert.Equal(t, 1, len(handler.Backups.ListObjects(ctx, "backups")))
	})

	t.Run("restore/success", func(t *testing.T) {
		ctx := context.Background()
		handler := BackupHandler{
			Database: &Database{Parts: newParts()},
			Backups:  newBucket(t),
		}
		backup, err := handler.Database.Backup(ctx)
		require.NoError(t, err, "database.Backup")

		backupJSON, err := json.Marshal(backup)
		require.NoError(t, err, "json.Marshal")
		require.NoError(t, handler.Backups.PutObject(ctx, "backup.json", &storage.Object{
			ContentType: "application/json",
			Bytes:       backupJSON,
		}), "bucket.PutObject")

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/", nil)
		request.Form = make(url.Values)
		request.Form.Add("cmd", "restore")
		request.Form.Add("object", "backup.json")
		handler.ServeHTTP(recorder, request)

		got := recorder.Result()
		assert.Equal(t, http.StatusOK, got.StatusCode)
	})

	t.Run("restore/no object", func(t *testing.T) {
		handler := BackupHandler{
			Database: &Database{Parts: newParts()},
			Backups:  newBucket(t),
		}

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/", nil)
		request.Form = make(url.Values)
		request.Form.Add("cmd", "restore")
		request.Form.Add("object", "")
		handler.ServeHTTP(recorder, request)

		got := recorder.Result()
		assert.Equal(t, http.StatusBadRequest, got.StatusCode)
	})

	t.Run("restore_from_file/success", func(t *testing.T) {
		ctx := context.Background()
		handler := BackupHandler{
			Database: &Database{Parts: newParts()},
			Backups:  newBucket(t),
		}
		backup, err := handler.Database.Backup(ctx)
		require.NoError(t, err, "database.Backup")

		backupJSON, err := json.Marshal(backup)
		require.NoError(t, err, "json.Marshal")

		// build the request body with a json file
		var body bytes.Buffer
		multipartWriter := multipart.NewWriter(&body)

		fileHeader := make(textproto.MIMEHeader)
		fileHeader.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "backup_file", "backup.json"))
		fileHeader.Set("Content-Type", "application/json")

		fileWriter, err := multipartWriter.CreatePart(fileHeader)
		require.NoError(t, err, "multipartWriter.CreateFormFile() failed")
		_, err = fileWriter.Write(backupJSON)
		require.NoError(t, err)
		require.NoError(t, multipartWriter.WriteField("cmd", "restore from file"), "multipartWriter.WriteField() failed")
		multipartWriter.Close()

		// make the request
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/", &body)
		request.Header.Set("Content-Type", multipartWriter.FormDataContentType())
		handler.ServeHTTP(recorder, request)

		got := recorder.Result()
		assert.Equal(t, http.StatusOK, got.StatusCode)
	})
}
