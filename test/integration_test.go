package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	dockerTypes "github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const vvgoPort = "42069"
const redisContainer = "vvgo_redis_1"

func assertEqualJSONObject(t *testing.T, want io.Reader, got io.Reader) {
	var wantObj map[string]interface{}
	var gotObj map[string]interface{}
	assert.NoError(t, json.NewDecoder(want).Decode(&wantObj))
	assert.NoError(t, json.NewDecoder(got).Decode(&gotObj))
	assert.Equal(t, wantObj, gotObj)
}

func TestVVGO(t *testing.T) {
	t.Log("Starting integration test")
	rand.Seed(time.Now().UnixNano())
	gitSha, _ := exec.Command("git", "rev-parse", "HEAD").Output()
	require.NotEmpty(t, gitSha)
	imageName := fmt.Sprintf("vvgo:%s", strings.TrimSpace(string(gitSha)))
	t.Log("Image:", imageName)
	containerName := fmt.Sprintf("vvgo-testing-%#x", rand.Uint64())
	t.Log("Container:", containerName)

	t.Run("Build Image", func(t *testing.T) {
		cmd := newCmd(t, "docker", "build",
			".",
			"--tag", "vvgo:latest",
			"--tag", imageName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		assert.NoError(t, runCmd(t, cmd))
	})

	t.Run("Test Image", func(t *testing.T) {
		require.NoError(t, runCmd(t, newCmd(t, "docker-compose", "up", "-d")))

		vvgoCmd := newCmd(t, "docker", "run",
			"--detach",
			"--rm",
			"--network", "vvgo_default",
			"--expose", vvgoPort,
			"--publish-all",
			"--env", "VVGO_LISTEN_ADDRESS=0.0.0.0:"+vvgoPort,
			"--env", "REDIS_ADDRESS="+redisContainer+":6379",
			"--name", containerName,
			imageName)
		require.NoError(t, runCmd(t, vvgoCmd))
		t.Cleanup(func() {
			cmd := newCmd(t, "docker", "stop", containerName)
			require.NoError(t, runCmd(t, cmd))
		})
		vvgoURL := "http://localhost:" + containerPort(t, containerName)

		t.Run("GET /", func(t *testing.T) {
			url := vvgoURL + "/"
			t.Log("GET", url)
			resp, err := http.Get(url)
			if assert.NoError(t, err) {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			}
		})

		t.Run("GET /api/v1/me", func(t *testing.T) {
			url := vvgoURL + "/api/v1/me"
			t.Log("GET", url)
			resp, err := http.Get(url)
			if assert.NoError(t, err) {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				assertEqualJSONObject(t,
					strings.NewReader(`{
						"Status":"ok",
						"Identity": {
							"Key": "",
							"Kind": "anonymous",
							"Roles": ["anonymous"],
							"ExpiresAt": "0001-01-01T00:00:00Z",
							"CreatedAt": "0001-01-01T00:00:00Z"
						}
					}`), resp.Body)
			}
		})

		t.Run("GET /api/nowhere", func(t *testing.T) {
			url := vvgoURL + "/api/nowhere"
			t.Log("GET", url)
			resp, err := http.Get(url)
			if assert.NoError(t, err) {
				assert.Equal(t, http.StatusNotFound, resp.StatusCode)
			}
		})

		t.Run("GET /nowhere", func(t *testing.T) {
			url := vvgoURL + "/nowhere"
			tempDir := t.TempDir()
			indexFileName := tempDir + "/index.html"
			cmd := newCmd(t, "docker", "cp", containerName+":/app/public/dist/index.html", indexFileName)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			require.NoError(t, runCmd(t, cmd))

			indexBytes, err := ioutil.ReadFile(indexFileName)
			require.NoError(t, err)

			t.Log("GET", url)
			resp, err := http.Get(url)
			if assert.NoError(t, err) {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				var body bytes.Buffer
				_, err := body.ReadFrom(resp.Body)
				assert.NoError(t, err)
				assert.Equal(t, string(indexBytes), body.String())
			}
		})
	})
}

func newCmd(t *testing.T, name string, args ...string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = findRoot(t)
	return cmd
}

func runCmd(t *testing.T, cmd *exec.Cmd) error {
	t.Helper()
	t.Log("Executing:", cmd)
	return cmd.Run()
}

func findRoot(t *testing.T) (workDir string) {
	t.Helper()
	for workDir, _ = os.Getwd(); workDir != filepath.Dir(workDir); workDir = filepath.Dir(workDir) {
		_, err := os.Stat(filepath.Join(workDir, "go.mod"))
		switch {
		case err == nil:
			return
		case os.IsNotExist(err):
			continue
		default:
			t.Fatal("os.Stat() failed:", err)
		}
	}
	t.Fatal("could not find go.mod")
	return
}

func inspectContainer(t *testing.T, name string) []dockerTypes.ContainerJSON {
	t.Helper()
	var resp []dockerTypes.ContainerJSON
	var buf bytes.Buffer
	inspectContainer := exec.Command("docker", "inspect", name)
	inspectContainer.Stderr = os.Stderr
	inspectContainer.Stdout = &buf
	require.NoError(t, inspectContainer.Run())
	require.NoError(t, json.NewDecoder(&buf).Decode(&resp))
	return resp
}

func containerPort(t *testing.T, name string) string {
	t.Helper()
	data := inspectContainer(t, name)
	require.NotEmpty(t, data, "no matches")
	require.Equal(t, 1, len(data), "unexpected matches")
	require.NotEmpty(t, data[0].NetworkSettings.Ports, "data[0].NetworkSettings.Ports")
	require.NotEmpty(t, data[0].NetworkSettings.Ports[vvgoPort+"/tcp"], fmt.Sprintf(`data[0].NetworkSettings.Ports["%s/tcp"]`, vvgoPort))
	return data[0].NetworkSettings.Ports[vvgoPort+"/tcp"][0].HostPort
}
