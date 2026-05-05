package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/google/gnostic/cmd/protoc-gen-openapi/generator"
	v3 "github.com/google/gnostic/openapiv3"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

var flags flag.FlagSet

// The code of this function is copied from https://github.com/google/gnostic/blob/main/cmd/protoc-gen-openapi/main.go
// commit ade94e0d08cb9c60272a311608cd5dabd30d1813.
// The original protoc-gen-openapi output is read and reprocessed with Flipt specifics.
func main() {
	conf := generator.Configuration{
		Version:         flags.String("version", "0.0.1", "version number text, e.g. 1.2.3"),
		Title:           flags.String("title", "", "name of the API"),
		Description:     flags.String("description", "", "description of the API"),
		Naming:          flags.String("naming", "json", `naming convention. Use "proto" for passing names directly from the proto files`),
		FQSchemaNaming:  flags.Bool("fq_schema_naming", false, `schema naming convention. If "true", generates fully-qualified schema names by prefixing them with the proto message package name`),
		EnumType:        flags.String("enum_type", "integer", `type for enum serialization. Use "string" for string-based serialization`),
		CircularDepth:   flags.Int("depth", 2, "depth of recursion for circular messages"),
		DefaultResponse: flags.Bool("default_response", true, `add default response. If "true", automatically adds a default response to operations which use the google.rpc.Status message. Useful if you use envoy or grpc-gateway to transcode as they use this type for their default error responses.`),
		OutputMode:      flags.String("output_mode", "merged", `output generation mode. By default, a single openapi.yaml is generated at the out folder. Use "source_relative' to generate a separate '[inputfile].openapi.yaml' next to each '[inputfile].proto'.`),
	}

	opts := protogen.Options{
		ParamFunc: flags.Set,
	}

	opts.Run(func(plugin *protogen.Plugin) error {
		// Enable "optional" keyword in front of type (e.g. optional string label = 1;)
		plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		if *conf.OutputMode == "source_relative" {
			for _, file := range plugin.Files {
				if !file.Generate {
					continue
				}
				outfileName := strings.TrimSuffix(file.Desc.Path(), filepath.Ext(file.Desc.Path())) + ".openapi.gen.out"
				outputFile := plugin.NewGeneratedFile(outfileName, "")
				gen := generator.NewOpenAPIv3Generator(plugin, conf, []*protogen.File{file})
				if err := gen.Run(outputFile); err != nil {
					return err
				}
				// modified original logic
				data, err := outputFile.Content()
				if err != nil {
					return err
				}
				outputFile.Skip() // don't store temporary output
				outfileName = strings.TrimSuffix(file.Desc.Path(), filepath.Ext(file.Desc.Path())) + ".openapi.yaml"
				outputFile = plugin.NewGeneratedFile(outfileName, "")
				return run(data, outputFile)
				// end of changes
			}
		} else {
			outputFile := plugin.NewGeneratedFile("openapi.gen.out", "")
			err := generator.NewOpenAPIv3Generator(plugin, conf, plugin.Files).Run(outputFile)
			if err != nil {
				return err
			}
			// modified original logic
			data, err := outputFile.Content()
			if err != nil {
				return err
			}
			outputFile.Skip() // don't store temporary output
			outputFile = plugin.NewGeneratedFile("openapi.yaml", "")
			return run(data, outputFile)
			// end of changes
		}
		return nil
	})
}

func run(data []byte, outputFile *protogen.GeneratedFile) error {
	doc, err := v3.ParseDocument(data)
	if err != nil {
		return err
	}

	for _, s := range doc.Components.Schemas.AdditionalProperties {
		if s.Name == "GoogleProtobufAny" {
			s.Value = &v3.SchemaOrReference{
				Oneof: &v3.SchemaOrReference_Schema{
					Schema: &v3.Schema{
						OneOf: []*v3.SchemaOrReference{
							{Oneof: &v3.SchemaOrReference_Reference{Reference: &v3.Reference{XRef: "#/components/schemas/FlagResourcePayload"}}},
							{Oneof: &v3.SchemaOrReference_Reference{Reference: &v3.Reference{XRef: "#/components/schemas/SegmentResourcePayload"}}},
						},
					},
				},
			}
		}
	}

	doc.Components.Schemas.AdditionalProperties = append(
		doc.Components.Schemas.AdditionalProperties,
		&v3.NamedSchemaOrReference{Name: "AtType", Value: buildPayloadType()},
		&v3.NamedSchemaOrReference{Name: "FlagResourcePayload", Value: buildAnyVariant("flipt.core.Flag", "Flag")},
		&v3.NamedSchemaOrReference{Name: "SegmentResourcePayload", Value: buildAnyVariant("flipt.core.Segment", "Segment")},
	)

	doc.Components.Schemas.AdditionalProperties = slices.DeleteFunc(
		doc.Components.Schemas.AdditionalProperties,
		func(e *v3.NamedSchemaOrReference) bool {
			return e.Name == "SchemaAnchor"
		},
	)

	doc.Paths.Path = slices.DeleteFunc(doc.Paths.Path, func(p *v3.NamedPathItem) bool {
		return p.Name == "/api/v2/_payloadschema"
	})

	bytes, err := doc.YAMLValue("# Generated with protoc-gen-flipt-openapi")
	if err != nil {
		return err
	}

	if _, err = outputFile.Write(bytes); err != nil {
		return fmt.Errorf("failed to write yaml: %s", err.Error())
	}
	return nil
}

func buildPayloadType() *v3.SchemaOrReference {
	return &v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Schema{
			Schema: &v3.Schema{
				Type:        "object",
				Description: "The type of the message.",
				Required:    []string{"@type"},
				Properties: &v3.Properties{
					AdditionalProperties: []*v3.NamedSchemaOrReference{
						{
							Name: "@type", Value: &v3.SchemaOrReference{
								Oneof: &v3.SchemaOrReference_Schema{
									Schema: &v3.Schema{
										Type:  "string",
										Title: "atType",
										Enum:  []*v3.Any{{Yaml: "flipt.core.Flag"}, {Yaml: "flipt.core.Segment"}},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func buildAnyVariant(typeURL string, schemaName string) *v3.SchemaOrReference {
	return &v3.SchemaOrReference{
		Oneof: &v3.SchemaOrReference_Schema{
			Schema: &v3.Schema{
				AllOf: []*v3.SchemaOrReference{
					{
						Oneof: &v3.SchemaOrReference_Reference{
							Reference: &v3.Reference{
								XRef: "#/components/schemas/AtType",
							},
						},
					},
					{
						Oneof: &v3.SchemaOrReference_Reference{
							Reference: &v3.Reference{
								XRef: "#/components/schemas/" + schemaName,
							},
						},
					},
				},
			},
		},
	}
}
