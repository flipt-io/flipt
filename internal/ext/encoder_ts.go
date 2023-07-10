package ext

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

type EncoderTS struct {
	writer io.Writer
}

func NewTSEncoder(w io.Writer) *EncoderTS {
	return &EncoderTS{
		writer: w,
	}
}

func (e *EncoderTS) Encode(v interface{}) error {
	document, ok := v.(*Document)
	if !ok {
		return fmt.Errorf("invalid document type")
	}

	// Generate Typescript code
	code := generateTypescriptCode(*document)

	// Write the code to the writer
	_, err := e.writer.Write([]byte(code))
	if err != nil {
		return err
	}

	return nil
}

func (e *EncoderTS) Close() error {
	if closer, ok := e.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func generateTypescriptCode(document Document) string {
	codeTemplate := `export type Namespaces = {
  {{ .Namespace }}: {
    Flags: {{ .FlagsType }};
    Context: {{ .ContextFields }};
  };
};`

	data := struct {
		Namespace     string
		FlagsType     string
		ContextFields string
	}{
		Namespace:     document.Namespace,
		FlagsType:     generateFlagsCode(document.Flags),
		ContextFields: generateContextCode(document.Segments),
	}

	tmpl := template.Must(template.New("typescriptCode").Parse(codeTemplate))

	var builder strings.Builder
	err := tmpl.Execute(&builder, data)
	if err != nil {
		fmt.Println("Error generating Typescript code:", err)
		return ""
	}

	return builder.String()
}

func generateFlagsCode(flags []*Flag) string {
	var builder strings.Builder

	for i, flag := range flags {
		variantKeys := make([]string, len(flag.Variants))
		for j, variant := range flag.Variants {
			variantKeys[j] = variant.Key
		}

		unionType := generateUnionType(variantKeys)

		builder.WriteString(fmt.Sprintf("{ key: '%s'; value: %s }", flag.Key, unionType))

		if i < len(flags)-1 {
			builder.WriteString(" | ")
		}
	}

	return builder.String()
}

func generateUnionType(keys []string) string {
	return "'" + strings.Join(keys, "' | '") + "'"
}

func generateContextCode(segments []*Segment) string {
	var builder strings.Builder

	for _, segment := range segments {
		for _, constraint := range segment.Constraints {
			fieldName := constraint.Property
			fieldType := determineFieldType(constraint.Type)

			fieldCode := fmt.Sprintf("%s: %s; ", fieldName, fieldType)
			builder.WriteString(fieldCode)
		}
	}

	if builder.Len() > 0 {
		return "{ " + builder.String() + "}"
	}

	return "never"
}

func determineFieldType(value string) string {
	switch value {
	case "NUMBER_COMPARISON_TYPE":
		return "number"
	case "BOOLEAN_COMPARISON_TYPE":
		return "boolean"
	default:
		return "string"
	}
}
