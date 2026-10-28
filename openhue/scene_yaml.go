package openhue

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseSceneDefinition decodes a YAML scene definition. Unknown keys and wrong types are errors that
// carry the line number.
func ParseSceneDefinition(r io.Reader) (*SceneDefinition, error) {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)

	def := &SceneDefinition{}
	if err := dec.Decode(def); err != nil {
		if errors.Is(err, io.EOF) {
			return def, nil
		}
		return nil, fmt.Errorf("invalid scene definition: %w", err)
	}
	if err := def.Validate(); err != nil {
		return nil, fmt.Errorf("invalid scene definition: %w", err)
	}
	return def, nil
}

// LoadSceneDefinition reads a definition from a file, or from the given reader when path is "-".
func LoadSceneDefinition(path string, stdin io.Reader) (*SceneDefinition, error) {
	if path == "-" {
		return ParseSceneDefinition(stdin)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read scene definition: %w", err)
	}
	defer func() { _ = f.Close() }()
	return ParseSceneDefinition(f)
}

// MarshalSceneDefinition encodes the definition as YAML.
func MarshalSceneDefinition(def *SceneDefinition) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(def); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
