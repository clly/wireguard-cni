package server

import (
	"context"
	"errors"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/hashicorp/go-uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	wireguardv1 "github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1"
)

const defaultPrefix = "10.0.0.0/8"

func Test_Register(t *testing.T) {
	testcases := []struct {
		name string
		req  *wireguardv1.RegisterRequest
		resp *wireguardv1.RegisterResponse
		err  error
	}{
		{
			name: "HappyPath",
			req: &wireguardv1.RegisterRequest{
				PublicKey: "abc123",
				Endpoint:  "192.168.1.1:51820",
				Route:     "10.0.0.9/24",
			},
			resp: &wireguardv1.RegisterResponse{},
		},
		{
			name: "MissingPK",
			req: &wireguardv1.RegisterRequest{
				PublicKey: "",
				Endpoint:  "192.168.1.1:51820",
				Route:     "10.0.0.0/24",
			},
			resp: nil,
			err:  connect.NewError(connect.CodeFailedPrecondition, errors.New("failed to validate: public_key: cannot be blank.")),
		},
		{
			name: "MissingEndpoint",
			req: &wireguardv1.RegisterRequest{
				PublicKey: "abc123",
				Endpoint:  "",
				Route:     "10.0.0.0/24",
			},
			resp: nil,
			err:  connect.NewError(connect.CodeFailedPrecondition, errors.New("failed to validate: endpoint: cannot be blank.")),
		},
		{
			name: "MissingRoute",
			req: &wireguardv1.RegisterRequest{
				PublicKey: "abc123",
				Endpoint:  "192.168.1.1:51820",
				Route:     "",
			},
			resp: nil,
			err:  connect.NewError(connect.CodeFailedPrecondition, errors.New("failed to validate: route: cannot be blank.")),
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			// d := t.TempDir()
			r := require.New(t)
			// jsonFile := filepath.Join(d, "wireguard.json")
			s, err := NewServer(defaultPrefix, CLUSTER_MODE, &WireguardServerConfig{
				// JSONOutputFile: jsonFile,
			})

			r.NoError(err)
			expectedResponse := connect.NewResponse(testcase.resp)
			req := connect.NewRequest(testcase.req)
			resp, err := s.Register(context.Background(), req)
			if testcase.err != nil {
				r.Error(err)
				r.EqualError(testcase.err, err.Error())
				// r.Eventuallyf(func() bool {
				// 	_, err := os.Stat(jsonFile)
				// 	return !os.IsNotExist(err)
				// }, 10*time.Second, 1*time.Second, "json state file does not exist")
			} else {
				r.Nil(err)
				r.Equal(expectedResponse, resp)
			}
		})
	}
}

func Test_Peers(t *testing.T) {
	testcases := []struct {
		name      string
		peersFunc func(t *testing.T, m *mapDB)
		err       error
	}{
		{
			name: "HappyPathEmpty",
		},
		{
			name: "HappyPathWithPeers",
			peersFunc: func(t *testing.T, m *mapDB) {
				r := validRegisterReq(t)
				b, err := protojson.Marshal(r)
				require.NoError(t, err)
				m.Set(r.PublicKey, string(b))
			},
		},
		/*
			{
				name: "BadValuesInMapDB",
				peersFunc: func(t *testing.T, m *mapDB) {
					m.Set("helo", "will-not-marshal")
				},
				err: connect.NewError(connect.CodeInternal, errors.New("proto: (line 1:2): unknown field \"will-not-marshal\""")),
			},
		*/
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			r := require.New(t)
			s, err := NewServer(defaultPrefix, CLUSTER_MODE, &WireguardServerConfig{})
			r.NoError(err)
			m, err := newMapDB()
			r.NoError(err)
			if testcase.peersFunc != nil {
				testcase.peersFunc(t, m)
				reqs := m.List()
				for _, reqS := range reqs {
					req, err := registerFromString(reqS)
					if err == nil {
						r.NoError(s.registerWGKey(req.PublicKey, req))
					} else {
						s.wgKey.Set("helo", `{"will-not-marshal":""}`)
					}
				}
			}

			req := connect.NewRequest(&wireguardv1.PeersRequest{})
			resp, err := s.Peers(context.Background(), req)
			if testcase.err != nil {
				r.Error(err)
				r.EqualError(err, testcase.err.Error())
			} else {
				r.NoError(err)
				for _, peer := range resp.Msg.GetPeers() {
					v, ok := m.Get(peer.PublicKey)
					r.True(ok)
					req, err := registerFromString(v)
					r.NoError(err)
					r.Equal(req.Endpoint, peer.Endpoint)
					r.Equal(req.PublicKey, peer.PublicKey)
					r.Equal(req.Route, peer.Route)
				}
			}
		})
	}
}

func self() *wireguardv1.Peer {
	return &wireguardv1.Peer{
		PublicKey: "abc123=",
		Endpoint:  "192.168.1.1:51820",
		Route:     "0.0.0.0/0",
	}
}

func setSelf(t *testing.T, m *mapDB) {
	r := require.New(t)
	p := self()
	req := &wireguardv1.RegisterRequest{
		PublicKey: p.PublicKey,
		Endpoint:  p.Endpoint,
		Route:     p.Route,
	}
	b, err := protojson.Marshal(req)
	r.NoError(err)
	m.Set(p.PublicKey, string(b))
}

func Test_PeersNodeMode(t *testing.T) {
	testcases := []struct {
		name      string
		peersFunc func(t *testing.T, m *mapDB)
		err       error
	}{
		{
			name: "HappyPathEmpty",
		},
		{
			name: "HappyPathWithPeers",
			peersFunc: func(t *testing.T, m *mapDB) {
				r := validRegisterReq(t)
				b, err := protojson.Marshal(r)
				require.NoError(t, err)
				m.Set(r.PublicKey, string(b))
			},
		},
		/*
			{
				name: "BadValuesInMapDB",
				peersFunc: func(t *testing.T, m *mapDB) {
					m.Set("helo", "will-not-marshal")
				},
				err: connect.NewError(connect.CodeInternal, errors.New("proto: (line 1:2): unknown field \"will-not-marshal\""")),
			},
		*/
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			r := require.New(t)
			p := self()
			s, err := NewServer(defaultPrefix, NODE_MODE, &WireguardServerConfig{Self: p})
			r.NoError(err)

			m, err := newMapDB()
			r.NoError(err)
			if testcase.peersFunc != nil {
				testcase.peersFunc(t, m)
				reqs := m.List()
				for _, reqS := range reqs {
					req, err := registerFromString(reqS)
					if err == nil {
						r.NoError(s.registerWGKey(req.PublicKey, req))
					} else {
						s.wgKey.Set("helo", `{"will-not-marshal":""}`)
					}
				}
			}
			setSelf(t, m)

			req := connect.NewRequest(&wireguardv1.PeersRequest{})
			resp, err := s.Peers(context.Background(), req)
			if testcase.err != nil {
				r.Error(err)
				r.EqualError(err, testcase.err.Error())
			} else {
				r.NoError(err)
				for _, peer := range resp.Msg.GetPeers() {
					v, ok := m.Get(peer.PublicKey)
					r.True(ok)
					req, err := registerFromString(v)
					r.NoError(err)
					r.Equal(req.Endpoint, peer.Endpoint)
					r.Equal(req.PublicKey, peer.PublicKey)
					r.Equal(req.Route, peer.Route)
				}
			}
		})
	}
}

func validRegisterReq(t *testing.T) *wireguardv1.RegisterRequest {
	u, err := uuid.GenerateUUID()
	require.NoError(t, err)
	r := &wireguardv1.RegisterRequest{
		PublicKey: u,
		Endpoint:  "192.168.1.1:51820",
		Route:     "10.0.0.0/24",
	}
	return r
}
