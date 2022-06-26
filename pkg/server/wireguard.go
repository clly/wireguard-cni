package server

import (
	"context"
	"fmt"
	wireguardv1 "wireguard-cni/gen/wgcni/wireguard/v1"
	"wireguard-cni/gen/wgcni/wireguard/v1/wireguardv1connect"

	"github.com/bufbuild/connect-go"
	validation "github.com/go-ozzo/ozzo-validation"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	_ wireguardv1connect.WireguardServiceHandler = &Server{}
)

func (s *Server) Register(ctx context.Context,
	req *connect.Request[wireguardv1.RegisterRequest],
) (*connect.Response[wireguardv1.RegisterResponse], error) {
	pk := req.Msg.GetPublicKey()

	err := validation.ValidateStruct(req.Msg,
		validation.Field(&req.Msg.Endpoint, validation.Required),
		validation.Field(&req.Msg.PublicKey, validation.Required),
		validation.Field(&req.Msg.Route, validation.Required),
	)

	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("failed to validate: %w", err))
	}

	err = s.registerWGKey(pk, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to register public key %w", err))
	}

	rp := &wireguardv1.RegisterResponse{}
	return connect.NewResponse(rp), nil
}

func (s *Server) registerWGKey(pk string, msg *wireguardv1.RegisterRequest) error {
	b, err := protojson.Marshal(msg)
	if err != nil {
		return fmt.Errorf("proto is broke %w", err)
	}
	s.wgKey.Set(pk, string(b))
	s.expvarMap.Set(pk, msg)
	return nil
}

func (s *Server) Peers(ctx context.Context,
	req *connect.Request[wireguardv1.PeersRequest],
) (*connect.Response[wireguardv1.PeersResponse], error) {

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

	p := &wireguardv1.PeersResponse{
		Peers: peers,
	}

	return connect.NewResponse(p), nil
}

func registerFromString(s string) (*wireguardv1.RegisterRequest, error) {
	p := &wireguardv1.RegisterRequest{}
	err := protojson.Unmarshal([]byte(s), p)
	if err != nil {
		return nil, err
	}

	return p, nil
}
