// xgface_service_restful project main.go
package main

import (
	"flag"
	"net/http"

	"github.com/Intelligentvision/faceAPI/config"
	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	gw "github.com/Intelligentvision/faceAPI/proto/faceAPI"
)

var (
	//echoEndpoint = flag.String("echo_endpoint", "localhost:10000", "endpoint of YourService")

	echoEndpoint = flag.String("echo_endpoint", config.Config.Services.Web.FaceAPIaddr, "endpoint of YourService")
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}))

	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := gw.RegisterXgfaceAPIServiceHandlerFromEndpoint(ctx, mux, *echoEndpoint, opts)
	if err != nil {
		return err
	}

	return http.ListenAndServe(":8099", mux)
}

func main() {
	flag.Parse()
	defer glog.Flush()

	if err := run(); err != nil {
		glog.Fatal(err)
	}
}
