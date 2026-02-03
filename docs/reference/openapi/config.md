# Configuration Reference

The OpenAPI 3.1 specification includes numerous fields and configurations, many of which are supported and recognized by the OpenAPI plug-in.

For comprehensive guidelines on OpenAPI 3.1, please refer [to this](https://spec.openapis.org/oas/v3.1.0.html).

## JSON Schema

We recommend using an IDE tool that supports JSON/YAML Schemas based on the file format. This can help you with auto-completion when using the configuration files.

??? Example
    For VSCode/Cursor, install [YAML extension](https://open-vsx.org/extension/redhat/vscode-yaml/0.9.1). Then, you can add on top of your configuration files:

    ```yaml
    # yaml-language-server: $schema=https://raw.githubusercontent.com/meshapi/grpc-api-gateway/refs/heads/main/api/Config.schema.json
    ```

### path_param_name

Allows you to specify an alternative name for this field when it is utilized as a path parameter.

### required

Allows you to mark the field as required on the field directly. See [Explicitly Define Required Fields](/grpc-api-gateway/reference/openapi/field_optionality#explicitly-define-required-fields) for other ways to mark the field as required.

## OpenAPI differences

While the majority of the annotations align closely with the OpenAPI specification, there are a few notable differences and unique elements that should be highlighted.

### Type Slice

Type in the OpenAPI 3.1 specification can be either a single string or an array of strings. In the annotations, it is always represented as an array. However, if only one item is present, the final OpenAPI output will simplify it to a single value.

### References

In Schemas, you can use `ref` to refer to a proto message and popuate its details.

=== "Proto file"

    ```proto linenums="1" hl_lines="15"
    import "gopencloud/gateway/annotations.proto";

    service MyService {
      rpc Do(Request) returns (Response) {
        option (gopencloud.gateway.openapi_operation) = {
          responses: [
            {
              key: "206",
              value: {
                content: [
                  {
                    key: "application/json",
                    value: {
                      schema: {
                        ref: ".main.PartialResponse"
                      }
                    }
                  }
                ]
              }
            }
          ]
        };
      };
   }
    ```

=== "Configuration"

    ```yaml linenums="1" hl_lines="11"
    openapi:
      services:
        - selector: "~.MyService"
          methods:
            Do:
              responses:
                "206":
                  content:
                    application/json:
                      schema:
                        ref: ".main.PartialResponse"
    ```

## Extras

If you need to specify any additional fields from the OpenAPI specification that are not covered by the annotations, or if you have custom fields you want to include (such as custom extensions), you can use the `extra` property to add these keys:

=== "Proto file"

    ```proto linenums="1" hl_lines="5-7"
    import "gopencloud/gateway/annotations.proto";

    message User {
      option (gopencloud.gateway.openapi_schema) = {
        extra: [
          {key: "x-custom-key", value: {string_value: "value}"}
        ]
      };
    }
    ```

=== "Configuration"

    ```yaml linenums="1" hl_lines="5-6"
    openapi:
      messages:
        - selector: "~.User"
          schema:
            extra:
              key: value
    ```
