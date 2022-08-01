package server

import (
	"expvar"
	"sync"

	"github.com/bufbuild/connect-go"
	wireguardv1 "github.com/clly/wireguard-cni/gen/wgcni/wireguard/v1"
	goipam "github.com/metal-stack/go-ipam"
)

type Server struct {
	wgKey     *mapDB
	expvarMap *expvar.Map
	prefix    *goipam.Prefix
	ipam      goipam.Ipamer
	mode      IPAM_MODE
	self      *wireguardv1.Peer
}

func (s *Server) ListPeers() ([]*wireguardv1.Peer, error) {
	keyList := s.wgKey.List()
	peers := make([]*wireguardv1.Peer, 0, len(keyList))
	for _, v := range keyList {
		regReq, err := registerFromString(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		p := &wireguardv1.Peer{
			PublicKey: regReq.GetPublicKey(),
			Endpoint:  regReq.GetEndpoint(),
			Route:     regReq.GetRoute(),
		}
		peers = append(peers, p)
	}
	return peers, nil
}

type mapDB struct {
	db map[string]string
	m  *sync.RWMutex
}

func newMapDB() *mapDB {
	return &mapDB{
		db: map[string]string{},
		m:  &sync.RWMutex{},
	}
}

func (m *mapDB) Set(k string, v string) {
	m.m.Lock()
	m.db[k] = v
	m.m.Unlock()
}

func (m *mapDB) Get(k string) (val string, ok bool) {
	m.m.RLock()
	val, ok = m.db[k]
	m.m.RUnlock()
	return val, ok
}

func (m *mapDB) List() []string {
	m.m.RLock()
	peers := make([]string, 0, len(m.db))
	for _, v := range m.db {
		peers = append(peers, v)
	}
	m.m.RUnlock()
	return peers
}
