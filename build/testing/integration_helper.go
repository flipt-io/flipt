package testing

import (
	"strings"
	
	"gopkg.in/yaml.v3"
)

// stripIDsFromYAML removes id fields from rollouts and constraints in YAML
func stripIDsFromYAML(yamlContent string) (string, error) {
	var doc map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &doc); err != nil {
		return "", err
	}

	// Strip IDs from flags' rollouts
	if flags, ok := doc["flags"].([]interface{}); ok {
		for _, flag := range flags {
			if f, ok := flag.(map[string]interface{}); ok {
				if rollouts, ok := f["rollouts"].([]interface{}); ok {
					for _, rollout := range rollouts {
						if r, ok := rollout.(map[string]interface{}); ok {
							delete(r, "id")
						}
					}
				}
			}
		}
	}

	// Strip IDs from segments' constraints
	if segments, ok := doc["segments"].([]interface{}); ok {
		for _, segment := range segments {
			if s, ok := segment.(map[string]interface{}); ok {
				if constraints, ok := s["constraints"].([]interface{}); ok {
					for _, constraint := range constraints {
						if c, ok := constraint.(map[string]interface{}); ok {
							delete(c, "id")
						}
					}
				}
			}
		}
	}

	// Marshal back to YAML
	output, err := yaml.Marshal(doc)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// normalizeYAMLForComparison removes IDs and normalizes YAML for comparison
func normalizeYAMLForComparison(yamlContent string) (string, error) {
	// Handle multi-document YAML (separated by ---)
	docs := strings.Split(yamlContent, "\n---\n")
	var normalizedDocs []string
	
	for _, doc := range docs {
		if strings.TrimSpace(doc) == "" {
			continue
		}
		normalized, err := stripIDsFromYAML(doc)
		if err != nil {
			return "", err
		}
		normalizedDocs = append(normalizedDocs, normalized)
	}
	
	return strings.Join(normalizedDocs, "---\n"), nil
}