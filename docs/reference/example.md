# Annotation example

Let's take a look at detailed example, covering most of annotation features.

```proto
syntax = "proto3";

package service.v1;

option go_package = ".;v1";

import "gopencloud/gateway/annotations.proto";

service ExampleAPI {
  rpc List(ListRequest) returns (ListResponse) {
    option (gopencloud.gateway.http) = {
      get: "/v1/list"
      query_params: [
        {selector: 'sort.key', name: 'sort_key'},
        {selector: 'sort.dir', name: 'sort_dir'},
        {selector: 'pagination.marker', name: 'marker'},
        {selector: 'pagination.limit', name: 'limit'},
        {selector: 'pagination.offset', name: 'offset'}
      ]
      // Will add additional_binding with method GET, route '/v1/show_all' and same query_params mapping
      aliases: ["/v1/show_all"]
    };
    option (gopencloud.gateway.openapi_operation) = {
      responses: [
        {
          key: "400";
          value: {
            ref: {
              uri: "#/components/responses/bad_request"
            }
          }
        }
      ]
      config: {
        default_response_code: "201"; // Generate default response with this code
        disable_default_error_response: true; // Skip default error response for this method
      }
    };
  }
}

message ListRequest {
  SortParams sort = 1;
  Pagination pagination = 2;
}

message ListResponse {
  repeated Something items = 1;
}

message SortParams {
  string key = 1;
  string dir = 2;
}

message Pagination {
  string marker = 1;
  uint32 limit = 2;
  uint32 offset = 3;
}
```
