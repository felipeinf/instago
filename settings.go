package ig

import (
	"encoding/json"
	"os"
)

// LoadSettingsFromFile reads and unmarshals a JSON session file produced by DumpSettingsToFile or DumpSettings.
func LoadSettingsFromFile(path string) (*Settings, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Settings
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	if s.Cookies == nil {
		s.Cookies = map[string]string{}
	}
	if s.AuthorizationData == nil {
		s.AuthorizationData = map[string]string{}
	}
	return &s, nil
}

// DumpSettingsToFile writes settings to path as indented JSON with restrictive file permissions.
func DumpSettingsToFile(path string, s *Settings) error {
	b, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}
