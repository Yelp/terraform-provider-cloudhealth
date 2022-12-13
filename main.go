package main

import (
	"cloudhealth/cloudhealth"
	goplugin "github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	tf5server "github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"google.golang.org/grpc"
)

// gRPC message limit of 64MB
const gRPCLimit = 64 << 20

func main() {
	// modified implementation of plugin.Serve() method from terraform SDK
	// this is done in order to increase the max GRPC limit from 4MB to 64MB
	serveConfig := goplugin.ServeConfig{
		HandshakeConfig: plugin.Handshake,
		GRPCServer: func(opts []grpc.ServerOption) *grpc.Server {
			return grpc.NewServer(append(opts,
				grpc.MaxSendMsgSize(gRPCLimit),
				grpc.MaxRecvMsgSize(gRPCLimit))...)
		},
		VersionedPlugins: map[int]goplugin.PluginSet{
			5: {
				"terraform-registry.yelpcorp.com/yelp/cloudhealth": &tf5server.GRPCProviderPlugin{
					GRPCProvider: func() tfprotov5.ProviderServer {
						return schema.NewGRPCProviderServer(cloudhealth.Provider())
					},
				},
			},
		},
	}
	goplugin.Serve(&serveConfig)
}
