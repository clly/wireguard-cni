package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	wireguardv1 "github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1"
)

func setupWireguardJSONFile(dataDir string, r *require.Assertions) func(r *require.Assertions) {
	peer1 := wireguardv1.RegisterRequest{
		PublicKey: "9jalV3EEBnVXahro0pRMQ+cHlmjE33Slo9tddzCVtCw=",
		Endpoint:  "192.0.2.103:51993",
		Route:     "10.0.0.2/32",
	}
	peer1Bytes, err := protojson.Marshal(&peer1)
	r.NoError(err)
	peer2 := wireguardv1.RegisterRequest{
		PublicKey: "2RzKFbGMx5g7fG0BrWCI7JIpGvcwGkqUaCoENYueJw4=",
		Endpoint:  "203.0.113.102:51902",
		Route:     "10.0.0.3/32",
	}
	peer2Bytes, err := protojson.Marshal(&peer2)
	r.NoError(err)
	wirguardJson := map[string]string{
		"9jalV3EEBnVXahro0pRMQ+cHlmjE33Slo9tddzCVtCw=": string(peer1Bytes),
		"2RzKFbGMx5g7fG0BrWCI7JIpGvcwGkqUaCoENYueJw4=": string(peer2Bytes),
	}

	f, err := os.OpenFile(filepath.Join(dataDir, nodeWireguardFile), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	r.NoError(err)
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "\t")
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(wirguardJson)
	r.NoError(err)

	return func(r *require.Assertions) {
		err := os.RemoveAll(filepath.Join(dataDir, nodeWireguardFile))
		r.NoError(err)
	}
}

func Test_WithJSONDB(t *testing.T) {
	t.Run("create new wireguard.json if not exists", func(t *testing.T) {
		r := require.New(t)
		dataDir := t.TempDir()
		teardown := setupWireguardJSONFile(dataDir, r)
		defer teardown(r)

		mapDBOpt := WithJSONDB(dataDir, nodeWireguardFile)
		_, err := newMapDB(mapDBOpt)
		r.NoError(err)
		r.FileExists(filepath.Join(dataDir, nodeWireguardFile))
	})
	t.Run("load wireguard keys if wireguard.json already exists", func(t *testing.T) {
		r := require.New(t)
		dataDir := t.TempDir()
		teardown := setupWireguardJSONFile(dataDir, r)
		defer teardown(r)

		mapDBOpt := WithJSONDB(dataDir, nodeWireguardFile)
		m, err := newMapDB(mapDBOpt)
		r.NoError(err)
		r.Equal(2, len(m.db))
	})

}
