// +build !nosdk

package sdktest

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"testing"

	"github.com/orlangure/gnomock/gnomockd"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	srv := http.Server{
		Addr:    "127.0.0.1:23042",
		Handler: gnomockd.Handler(),
	}

	defer func() {
		_ = srv.Shutdown(context.Background())
	}()

	go func() {
		_ = srv.ListenAndServe()
	}()

	return m.Run()
}

func TestPython(t *testing.T) {
	t.Parallel()

	require.NoError(t, os.Chdir("./python"))

	cmd := exec.Command("/bin/bash", "./run.sh")
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "got error with output: %s", string(out))
}
