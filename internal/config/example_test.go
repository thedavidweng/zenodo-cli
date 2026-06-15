package config

import (
	"fmt"
	"os"
)

func ExampleCredentialsFromProfileAndEnv() {
	// Clear env vars for deterministic output.
	for _, k := range []string{"ZENODO_TOKEN", "ZENODO_SANDBOX", "ZENODO_API_URL"} {
		_ = os.Unsetenv(k)
	}

	p := &Profile{
		Token:   "my-token",
		Sandbox: true,
	}
	creds := CredentialsFromProfileAndEnv(p)
	fmt.Println("Token:", creds.Token)
	fmt.Println("Sandbox:", creds.Sandbox)
	fmt.Println("BaseURL:", creds.BaseURL)
	// Output:
	// Token: my-token
	// Sandbox: true
	// BaseURL: https://sandbox.zenodo.org
}

func ExampleLoad() {
	// Create a temporary config file.
	tmp, err := os.CreateTemp("", "example-config-*.yml")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer func() { _ = os.Remove(tmp.Name()) }()

	content := `current_profile: default
profiles:
  default:
    token: test-token
    sandbox: true
`
	if _, err := tmp.WriteString(content); err != nil {
		fmt.Println("Error:", err)
		return
	}
	_ = tmp.Close()

	cfg, err := Load(tmp.Name())
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Profile:", cfg.CurrentProfile)
	p, _ := cfg.GetProfile("default")
	fmt.Println("Token:", p.Token)
	fmt.Println("Sandbox:", p.Sandbox)
}
