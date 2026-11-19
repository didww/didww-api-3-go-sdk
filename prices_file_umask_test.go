//go:build linux || darwin

package didww

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPriceFileDownloadToFileFollowsUmask(t *testing.T) {
	gzipFixture := loadFixture(t, "prices/voice_in_prices.csv.gz")
	_, client := newPricesTestServer(t, func(_ int, w http.ResponseWriter, r *http.Request) {
		writePriceFile(w, r, gzipFixture, lastModifiedHeader)
	})
	oldUmask := syscall.Umask(0o027)
	t.Cleanup(func() { syscall.Umask(oldUmask) })
	dir := t.TempDir()
	newPath := filepath.Join(dir, "voice_in_prices.csv.gz")
	oldPath := filepath.Join(dir, "voice_out_prices.csv.gz")
	require.NoError(t, os.WriteFile(oldPath, []byte("old copy"), 0o600))

	_, err := client.VoiceInPrices().DownloadToFile(context.Background(), newPath)
	require.NoError(t, err)
	_, err = client.VoiceInPrices().DownloadToFile(context.Background(), oldPath)
	require.NoError(t, err)

	assert.Equal(t, os.FileMode(0o640), fileMode(t, newPath))
	assert.Equal(t, os.FileMode(0o640), fileMode(t, oldPath))
}
