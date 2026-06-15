package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func FuzzConfigLoad(f *testing.F) {
	f.Add([]byte(`current_profile: prod
profiles:
  prod:
    token: "tok123"
    sandbox: false
    base_url: "https://zenodo.org/api"
    endpoints:
      api: "https://zenodo.org/api"
`))
	f.Add([]byte(`current_profile: sandbox
profiles:
  sandbox:
    token: "test-token"
    sandbox: true
`))
	f.Add([]byte(`{}`))
	f.Add([]byte(``))
	f.Add([]byte(`  - :\n\tbad: [}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var cfg Config
		_ = yaml.Unmarshal(data, &cfg)
	})
}
