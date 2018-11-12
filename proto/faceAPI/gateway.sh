#!/bin/bash

protoc -I. faceAPI.proto --go_out=plugins=grpc:. 
protoc -I.  --grpc-gateway_out=logtostderr=true,grpc_api_configuration=faceAPI.yaml:. faceAPI.proto