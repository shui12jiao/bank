package apig

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	// map[grpcgateway-accept:[*/*] grpcgateway-content-type:[application/json] grpcgateway-user-agent:[Thunder Client (https://www.thunderclient.com)] x-forwarded-for:[127.0.0.1] x-forwarded-host:[localhost:8080]]	grpcGatewayUserAgentHeader   = "grpcgateway-user-agent"
	grpcGatewayContentTypeHeader = "grpcgateway-content-type"
	grpcGatewayUserAgentHeader   = "grpcgateway-user-agent"
	xForwardedForHeader          = "x-forwarded-for"
	xForwardedHostHeader         = "x-forwarded-host"

	// md: map[:authority:[localhost:8081] content-type:[application/grpc] grpc-client:[evans] user-agent:[grpc-go/1.48.0]]
	authorityHeader   = "authority"
	contentTypeHeader = "content-type"
	grpcClientHeader  = "grpc-client"
	userAgentHeader   = "user-agent"
)

const (
	// ContentTypeJSON is the content type for JSON
	ContentTypeJSON = "application/json"
	// ContentTypeGRPC is the content type for GRPC
	ContentTypeGRPC = "application/grpc"
)

type Metadata struct {
	UserAgent   string
	ClientIP    string
	ContentType string
}

func (server *Server) extractMetadata(ctx context.Context) *Metadata {
	mtdt := &Metadata{}
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		// log.Printf("md: %v", md)
		if contentTypes := md.Get(contentTypeHeader); len(contentTypes) > 0 {
			mtdt.ContentType = contentTypes[0]
		} else if contentTypes := md.Get(grpcGatewayContentTypeHeader); len(contentTypes) > 0 {
			mtdt.ContentType = contentTypes[0]
		}

		switch mtdt.ContentType {
		case ContentTypeJSON:
			if userAgents := md.Get(grpcGatewayUserAgentHeader); len(userAgents) > 0 {
				mtdt.UserAgent = userAgents[0]
			}
			if clientIPs := md.Get(xForwardedForHeader); len(clientIPs) > 0 {
				mtdt.ClientIP = clientIPs[0]
			}
		case ContentTypeGRPC:
			if userAgents := md.Get(userAgentHeader); len(userAgents) > 0 {
				mtdt.UserAgent = userAgents[0]
			}
			if p, ok := peer.FromContext(ctx); ok {
				mtdt.ClientIP = p.Addr.String()
			}
		}
	}
	return mtdt
}
