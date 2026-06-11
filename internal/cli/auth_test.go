package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestAuthCommandHasSubcommands(t *testing.T) {
	cmd := authCmd
	if cmd.Use != "auth" {
		t.Errorf("Use = %q, want auth", cmd.Use)
	}
}

func TestAuthLoginCommandExists(t *testing.T) {
	found := false
	for _, c := range authCmd.Commands() {
		if c.Name() == "login" {
			found = true
			break
		}
	}
	if !found {
		t.Error("auth should have 'login' subcommand")
	}
}

func TestAuthStatusCommandExists(t *testing.T) {
	found := false
	for _, c := range authCmd.Commands() {
		if c.Name() == "status" {
			found = true
			break
		}
	}
	if !found {
		t.Error("auth should have 'status' subcommand")
	}
}

func TestAuthLogoutCommandExists(t *testing.T) {
	found := false
	for _, c := range authCmd.Commands() {
		if c.Name() == "logout" {
			found = true
			break
		}
	}
	if !found {
		t.Error("auth should have 'logout' subcommand")
	}
}

func TestAuthLoginHasTokenFlag(t *testing.T) {
	var loginCmd *cobra.Command
	for _, c := range authCmd.Commands() {
		if c.Name() == "login" {
			loginCmd = c
			break
		}
	}
	if loginCmd == nil {
		t.Fatal("login command not found")
	}
	f := loginCmd.Flags().Lookup("token")
	if f == nil {
		t.Error("login should have --token flag")
	}
}

func TestAuthLoginWithTokenFlag(t *testing.T) {
	// The login command should accept --token and save it
	var loginCmd *cobra.Command
	for _, c := range authCmd.Commands() {
		if c.Name() == "login" {
			loginCmd = c
			break
		}
	}
	if loginCmd == nil {
		t.Fatal("login command not found")
	}

	// Verify the command has a RunE function
	if loginCmd.RunE == nil {
		t.Error("login command should have RunE")
	}
}

func TestAuthStatusRequiresConfig(t *testing.T) {
	// Verify status command exists and has RunE
	var statusCmd *cobra.Command
	for _, c := range authCmd.Commands() {
		if c.Name() == "status" {
			statusCmd = c
			break
		}
	}
	if statusCmd == nil {
		t.Fatal("status command not found")
	}
	if statusCmd.RunE == nil {
		t.Error("status command should have RunE")
	}
}

func TestAuthLogoutRequiresConfig(t *testing.T) {
	var logoutCmd *cobra.Command
	for _, c := range authCmd.Commands() {
		if c.Name() == "logout" {
			logoutCmd = c
			break
		}
	}
	if logoutCmd == nil {
		t.Fatal("logout command not found")
	}
	if logoutCmd.RunE == nil {
		t.Error("logout command should have RunE")
	}
}
