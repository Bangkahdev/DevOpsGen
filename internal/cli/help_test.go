package cli

import (
	"testing"
)

func TestNewRootCommand_Help(t *testing.T) {
	root := NewRootCommand()
	root.SetArgs([]string{"--help"})
	err := root.Execute()
	if err != nil {
		t.Fatalf("help: %v", err)
	}
}

func TestInitCommand_Help(t *testing.T) {
	cmd := newInitCommand()
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("init help: %v", err)
	}
}

func TestDetectCommand_Help(t *testing.T) {
	cmd := newDetectCommand()
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("detect help: %v", err)
	}
}

func TestValidateCommand_Help(t *testing.T) {
	cmd := newValidateCommand()
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("validate help: %v", err)
	}
}

func TestListCommand_Executes(t *testing.T) {
	cmd := newListCommand()
	if err := cmd.Flags().Parse(nil); err != nil {
		t.Fatal(err)
	}
}

func TestVersionCommand_Executes(t *testing.T) {
	cmd := newVersionCommand()
	if err := cmd.Flags().Parse(nil); err != nil {
		t.Fatal(err)
	}
}
