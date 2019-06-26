package internal

import (
	"errors"

	"github.com/solution9th/S3Adapter/internal/auth"
	"github.com/solution9th/S3Adapter/internal/gateway"
	"github.com/solution9th/S3Adapter/internal/gateway/cos"
	"github.com/solution9th/S3Adapter/internal/gateway/s3"

	"github.com/haozibi/zlog"
)

var (
	// ErrGatewayNotFound gateway not found
	ErrGatewayNotFound = errors.New("gateway not found")
)

var (
	// GatewayMap gateway map structure
	GatewayMap map[string]func() gateway.Gateway
)

func init() {
	GatewayMap = make(map[string]func() gateway.Gateway)
	GatewayMap[s3.Backend] = s3.New
	GatewayMap[cos.Backend] = cos.New
}

// NewGateway new gateway by accessKey, secretKey and region
func NewGateway(gatewayName string, creds auth.Credentials, region string) (pro gateway.S3Protocol, err error) {

	gf, ok := GatewayMap[gatewayName]
	if !ok {
		return nil, ErrGatewayNotFound
	}

	g := gf()

	if !g.Production() {
		return pro, ErrGatewayNotFound
	}

	zlog.ZInfo().Str("gateway", g.Name()).Str("region", region).Msg("[core]")

	return g.NewS3Protocol(creds, region, true)
}
