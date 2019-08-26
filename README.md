
# Prerequisite

    docker-compose up -d --build

# Server

Build server

    cd cmd/server/ && go build .


Run server

    ./server -grpc-port=9000 -db-host=127.0.0.1:3306 -db-user=root -db-password=root -db-schema=grpc

# Client

Build client

    cd cmd/client-grpc/ && go build .


Run client

    ./client-grpc -server=127.0.0.1:9000


# Update Protobuf

Edit `api/proto/v1/todo-service.proto`

Compile:

    protoc --proto_path=api/proto/v1 --proto_path=third_party --go_out=plugins=grpc:pkg/api/v1 todo-service.proto
