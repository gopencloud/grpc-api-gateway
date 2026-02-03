# Create a Simple Service

In this guide, we will generate and implement a simple gRPC service. In the following section, we will add the reverse proxy and the HTTP bindings.

## Echo Service

To get started, let's create the following proto file.

```proto title="echo_service.proto" linenums="1"
syntax = "proto3";

package echo;

option go_package = "demo/echo";

message EchoRequest {
    string text = 1;
    bool capitalize = 2;
}

message EchoResponse {
    string text = 1;
}

service EchoService {
    // Echo returns the received text and make it louder too!
    rpc Echo(EchoRequest) returns (EchoResponse);
}
```

## Code Generation

[EasyP](https://easyp.tech) is a tool that simplifies the development and consumption of Protobuf APIs.
One of its features is managing dependencies and building proto files.

If you decide to use EasyP, follow the instructions below or switch to the `protoc` tab for instructions using protoc.

Let's create a easyp.yaml with the following content:

=== "Using Buf"

    ```yaml title="easyp.yaml" linenums="1"
    deps:
      - github.com/gopencloud/grpc-api-gateway
    generate:
      inputs:
        - directory: .
    plugins:
      - name: go
        out: gen

      - name: go-grpc
        out: gen

      - name: grpc-api-gateway
        out: gen

      - name: openapiv3
        out: gen
    ```

    Now generate the artifacts using:

    ```sh
    easyp g
    ```

    You should see the generated files inside the `gen` directory.

=== "Using protoc"

    ```sh
    protoc \
        --go_out=gen \
        --go-grpc_out=gen \
        --grpc-api-gateway_out=gen \
        --openapiv3_out=gen \
        echo_service.proto
    ```

    You should see the generated files inside the `gen` directory.

## Implementing the Service

First, let's set up our Go module:

```sh
go mod init demo
```

The following `main.go` file implements the Echo service and starts a gRPC server on port `40000`:

```go title="main.go" linenums="1"
package main

import (
    "context"
    "demo/gen/demo/echo"
    "log"
    "net"
    "strings"

    "google.golang.org/grpc"
)

type Service struct {
    echo.UnimplementedEchoServiceServer
}

func (Service) Echo(
    ctx context.Context,
    req *echo.EchoRequest,
) (*echo.EchoResponse, error) {
    response := &echo.EchoResponse{
        Text: req.Text,
    }

    if req.Capitalize {
        response.Text = strings.ToUpper(response.Text)
    }

    return response, nil
}

func main() {
    listener, err := net.Listen("tcp", ":40000")
    if err != nil {
        log.Fatalf("failed to listen: %s", err)
    }

    server := grpc.NewServer()
    echo.RegisterEchoServiceServer(server, Service{})

    if err := server.Serve(listener); err != nil {
        log.Fatalf("gRPC server failed: %v", err)
    }
}
```

## Running the Service

Let's run it and ensure everything works correctly:

```sh
go mod tidy
go run .
```

If everything looks good, let's proceed to the next part and add the HTTP bindings!
