package main

import (
	"github.com/Intelligentvision/faceAPI/config"
	"github.com/Intelligentvision/faceAPI/proto/faceAPI"
	"github.com/Intelligentvision/faceAPI/server"
	//_ "github.com/Intelligentvision/StaffRepository/server/model"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", config.Config.Services.Web.Baseip+":"+config.Config.Services.Web.Port)
	if err != nil {
		log.Fatalf("failed to listen： %v", err)
	}
	s := grpc.NewServer()
	xgfaceAPI.RegisterXgfaceAPIServiceServer(s, &server.FaceAPIServer{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
