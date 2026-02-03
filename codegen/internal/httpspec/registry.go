package httpspec

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/gopencloud/grpc-api-gateway/api"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"gopkg.in/yaml.v3"
)

var (
	selectorPattern = regexp.MustCompile(`^\w+(?:[.]\w+)+$`)
)

type SourceInfo struct {
	Filename     string
	ProtoPackage string
}

type EndpointSpec struct {
	Binding    *api.EndpointBinding
	SourceInfo SourceInfo
}

type Registry struct {
	endpoints map[string]EndpointSpec
}

func NewRegistry() *Registry {
	return &Registry{endpoints: map[string]EndpointSpec{}}
}

// LoadFromFile loads a gateway config file for a proto file at filePath. if protoPackage is provided
// it will be used to convert relative selectors to absolute selectors.
func (r *Registry) LoadFromFile(filePath, protoPackage string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	defer file.Close()

	ctx := SourceInfo{
		Filename:     filePath,
		ProtoPackage: protoPackage,
	}

	switch filepath.Ext(filePath) {
	case ".json":
		return r.loadJSON(file, ctx)
	case ".yaml", ".yml":
		return r.loadYAML(file, ctx)
	default:
		return fmt.Errorf("unrecognized/unsupported file extension: %s", filePath)
	}
}

// Iterate iterates over all registered endpoint specifications.
func (r *Registry) Iterate(cb func(fqmn string, spec EndpointSpec)) {
	for key, spec := range r.endpoints {
		cb(key, spec)
	}
}

func (r *Registry) LoadFromService(filePath, protoPackage string, service *descriptorpb.ServiceDescriptorProto) error {
	config := api.Config{
		Gateway: &api.GatewaySpec{
			Endpoints: []*api.EndpointBinding{},
		},
	}

	for _, method := range service.GetMethod() {
		embeddedBinding, ok := proto.GetExtension(method.Options, api.E_Http).(*api.ProtoEndpointBinding)
		if !ok || embeddedBinding == nil {
			continue
		}

		endpointBinding := &api.EndpointBinding{
			Selector:                   protoPackage + "." + service.GetName() + "." + method.GetName(),
			Body:                       embeddedBinding.GetBody(),
			QueryParams:                embeddedBinding.GetQueryParams(),
			AdditionalBindings:         embeddedBinding.GetAdditionalBindings(),
			DisableQueryParamDiscovery: embeddedBinding.DisableQueryParamDiscovery,
			Stream:                     embeddedBinding.Stream,
			Aliases:                    embeddedBinding.GetAliases(),
		}
		setPatternFromProtoDefinition(embeddedBinding.Pattern, endpointBinding)
		config.Gateway.Endpoints = append(config.Gateway.Endpoints, endpointBinding)
	}

	return r.processConfig(&config, SourceInfo{Filename: filePath, ProtoPackage: protoPackage})
}

func setPatternFromProtoDefinition(value interface{}, binding *api.EndpointBinding) {
	switch value := value.(type) {
	case *api.ProtoEndpointBinding_Get:
		binding.Pattern = &api.EndpointBinding_Get{Get: value.Get}
	case *api.ProtoEndpointBinding_Put:
		binding.Pattern = &api.EndpointBinding_Put{Put: value.Put}
	case *api.ProtoEndpointBinding_Post:
		binding.Pattern = &api.EndpointBinding_Post{Post: value.Post}
	case *api.ProtoEndpointBinding_Delete:
		binding.Pattern = &api.EndpointBinding_Delete{Delete: value.Delete}
	case *api.ProtoEndpointBinding_Patch:
		binding.Pattern = &api.EndpointBinding_Patch{Patch: value.Patch}
	case *api.ProtoEndpointBinding_Custom:
		binding.Pattern = &api.EndpointBinding_Custom{Custom: value.Custom}
	}
}

func (r *Registry) loadYAML(reader io.Reader, src SourceInfo) error {
	var yamlContents interface{}
	if err := yaml.NewDecoder(reader).Decode(&yamlContents); err != nil {
		return fmt.Errorf("failed to decode yaml: %w", err)
	}

	jsonContents, err := json.Marshal(yamlContents)
	if err != nil {
		return fmt.Errorf("failed to JSON marshal content: %w", err)
	}

	config := &api.Config{}
	if err := protojson.Unmarshal(jsonContents, config); err != nil {
		return err
	}

	return r.processConfig(config, src)
}

func (r *Registry) loadJSON(reader io.Reader, src SourceInfo) error {
	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read configuration: %w", err)
	}

	config := &api.Config{}
	if err := protojson.Unmarshal(content, config); err != nil {
		return fmt.Errorf("failed to unmarshal json file: %s", err)
	}

	return r.processConfig(config, src)
}

func (r *Registry) processConfig(config *api.Config, src SourceInfo) error {
	if config.Gateway == nil {
		return nil
	}

	for _, endpoint := range config.Gateway.GetEndpoints() {
		// If selector starts with '.', it indicates it is relative to the proto package.
		if strings.HasPrefix(endpoint.Selector, "~.") {
			if src.ProtoPackage == "" {
				return fmt.Errorf("no proto package context is available, cannot use relative selector: %s", endpoint.Selector)
			}
			endpoint.Selector = src.ProtoPackage + endpoint.Selector[1:]
		}

		endpoint.Aliases = filterAliases(
			endpoint.GetAdditionalBindings(),
			endpoint.GetAliases(),
		)

		endpoint.AdditionalBindings = append(
			endpoint.AdditionalBindings,
			aliasesToAdditionalBindings(endpoint)...,
		)

		if err := validateBinding(endpoint); err != nil {
			return err
		}

		// add a leading dot to the selector to make it easy to look up FQMNs.
		if !strings.HasPrefix(endpoint.Selector, ".") {
			endpoint.Selector = "." + endpoint.Selector
		}

		if existingBinding, ok := r.endpoints[endpoint.Selector]; ok {
			return fmt.Errorf(
				"conflicting binding for %q: both %q and %q contain bindings for this selector",
				endpoint.Selector, src.Filename, existingBinding.SourceInfo.Filename)
		}

		r.endpoints[endpoint.Selector] = EndpointSpec{
			Binding:    endpoint,
			SourceInfo: src,
		}
	}

	return nil
}

// LookupBinding looks up endpoint bindings for a service method via a selector which is a FQMN.
func (r *Registry) LookupBinding(selector string) (EndpointSpec, bool) {
	result, ok := r.endpoints[selector]
	return result, ok
}

func validateBinding(endpoint *api.EndpointBinding) error {
	if !selectorPattern.MatchString(endpoint.Selector) {
		return fmt.Errorf("invalid selector: %q", endpoint.Selector)
	}

	if endpoint.Body != "" && endpoint.Body != "*" && !selectorPattern.MatchString(endpoint.Selector) {
		return fmt.Errorf("invalid body selector for %q: %s", endpoint.Selector, endpoint.Body)
	}

	for _, binding := range endpoint.AdditionalBindings {
		if binding.Body != "" && binding.Body != "*" && !selectorPattern.MatchString(endpoint.Selector) {
			return fmt.Errorf("invalid body selector %q: %s", endpoint.Selector, endpoint.Body)
		}
	}

	return nil
}

func aliasesToAdditionalBindings(e *api.EndpointBinding) []*api.AdditionalEndpointBinding {
	pattern := e.GetPattern()
	if pattern == nil {
		return nil
	}

	aliases := e.GetAliases()
	if len(aliases) == 0 {
		return nil
	}

	newBindings := make([]*api.AdditionalEndpointBinding, 0, len(aliases))
	for _, alias := range aliases {
		newBinding := &api.AdditionalEndpointBinding{}

		switch p := pattern.(type) {
		case *api.EndpointBinding_Get:
			newBinding.Pattern = &api.AdditionalEndpointBinding_Get{
				Get: alias,
			}
		case *api.EndpointBinding_Put:
			newBinding.Pattern = &api.AdditionalEndpointBinding_Put{
				Put: alias,
			}
		case *api.EndpointBinding_Post:
			newBinding.Pattern = &api.AdditionalEndpointBinding_Post{
				Post: alias,
			}
		case *api.EndpointBinding_Delete:
			newBinding.Pattern = &api.AdditionalEndpointBinding_Delete{
				Delete: alias,
			}
		case *api.EndpointBinding_Patch:
			newBinding.Pattern = &api.AdditionalEndpointBinding_Patch{
				Patch: alias,
			}
		case *api.EndpointBinding_Custom:
			newBinding.Pattern = &api.AdditionalEndpointBinding_Custom{
				Custom: &api.CustomPattern{
					Method: p.Custom.Method,
					Path:   alias,
				},
			}
		default:
			return nil
		}

		newBinding.Body = e.Body
		newBinding.ResponseBody = e.ResponseBody
		newBinding.QueryParams = e.QueryParams
		newBinding.DisableQueryParamDiscovery = e.DisableQueryParamDiscovery
		newBinding.Stream = e.Stream

		newBindings = append(newBindings, newBinding)
	}

	return newBindings
}

func filterAliases(abs []*api.AdditionalEndpointBinding, aliases []string) []string {
	if len(aliases) == 0 {
		return nil
	}

	if len(abs) == 0 {
		return aliases
	}

	existingPaths := make([]string, 0)
	for _, ab := range abs {
		abPattern := ab.GetPattern()
		if abPattern == nil {
			continue
		}

		path := ""
		switch p := abPattern.(type) {
		case *api.AdditionalEndpointBinding_Get:
			path = p.Get
		case *api.AdditionalEndpointBinding_Put:
			path = p.Put
		case *api.AdditionalEndpointBinding_Post:
			path = p.Post
		case *api.AdditionalEndpointBinding_Delete:
			path = p.Delete
		case *api.AdditionalEndpointBinding_Patch:
			path = p.Patch
		case *api.AdditionalEndpointBinding_Custom:
			path = p.Custom.Path
		}
		if path == "" {
			continue
		}

		if slices.Contains(aliases, path) {
			existingPaths = append(existingPaths, path)
		}
	}

	if len(existingPaths) == 0 {
		return aliases
	}

	newAliases := make([]string, 0, len(aliases))
	for _, alias := range aliases {
		if !slices.Contains(existingPaths, alias) {
			newAliases = append(newAliases, alias)
		}
	}

	return newAliases
}
