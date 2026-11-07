package skill

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"openhue-cli/openhue"
)

const (
	// Name is the folder the skill is installed into, matching the `name` in SKILL.md
	Name     = "openhue"
	FileName = "SKILL.md"

	docShortInstall = `Install the openhue skill for Claude Code`
	docLongInstall  = `
Install the openhue skill so that Claude Code (and other agents supporting skills) knows how to use the openhue CLI.

The skill is bundled in this binary, so it always matches the installed version of the CLI.
Run the command again after upgrading openhue to refresh it.

By default, the skill is installed for the current user in ~/.claude/skills/openhue/SKILL.md
($CLAUDE_CONFIG_DIR/skills when that variable is set).`
	docExampleInstall = `
# Install the skill for the current user
openhue install-skill

# Install the skill in the current project only (./.claude/skills)
openhue install-skill --project

# Install the skill in a custom skills directory
openhue install-skill --dir ~/my-agent/skills

# Print the skill instead of installing it
openhue install-skill --print`
)

type Options struct {
	dir     string
	project bool
	print   bool
}

// NewCmdInstallSkill creates the install-skill command that writes the given skill content on disk
func NewCmdInstallSkill(io openhue.IOStreams, content []byte) *cobra.Command {

	o := Options{}

	cmd := &cobra.Command{
		Use:     "install-skill",
		GroupID: "config",
		Short:   docShortInstall,
		Long:    docLongInstall,
		Example: docExampleInstall,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			if o.print {
				_, err := io.Out.Write(content)
				return err
			}

			dir, err := o.skillsDir()
			if err != nil {
				return err
			}

			path, status, err := Install(content, dir)
			if err != nil {
				return err
			}

			io.Printf("[OK] Skill %s in %s\n", status, path)
			return nil
		},
	}

	cmd.Flags().StringVar(&o.dir, "dir", "", "Skills directory to install into (the skill goes in <dir>/openhue/SKILL.md)")
	cmd.Flags().BoolVar(&o.project, "project", false, "Install in the current project (./.claude/skills) instead of for the current user")
	cmd.Flags().BoolVar(&o.print, "print", false, "Print the skill to stdout instead of installing it")
	cmd.MarkFlagsMutuallyExclusive("dir", "project", "print")

	return cmd
}

// skillsDir resolves the skills directory from the options
func (o *Options) skillsDir() (string, error) {
	switch {
	case o.dir != "":
		return o.dir, nil
	case o.project:
		return filepath.Join(".claude", "skills"), nil
	default:
		return UserSkillsDir()
	}
}

// UserSkillsDir returns the Claude Code user skills directory
func UserSkillsDir() (string, error) {
	if configDir := os.Getenv("CLAUDE_CONFIG_DIR"); configDir != "" {
		return filepath.Join(configDir, "skills"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to find the user home directory: %w", err)
	}

	return filepath.Join(home, ".claude", "skills"), nil
}

// Install writes the skill content in <skillsDir>/openhue/SKILL.md.
// It returns the file path and whether the skill was "installed", "updated" or is "already up to date".
func Install(content []byte, skillsDir string) (string, string, error) {

	if len(content) == 0 {
		return "", "", errors.New("this openhue binary does not contain the skill")
	}

	path := filepath.Join(skillsDir, Name, FileName)

	status := "installed"
	existing, err := os.ReadFile(path)
	switch {
	case err == nil && bytes.Equal(existing, content):
		return path, "already up to date", nil
	case err == nil:
		status = "updated"
	case !errors.Is(err, os.ErrNotExist):
		return "", "", fmt.Errorf("unable to read %s: %w", path, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", "", fmt.Errorf("unable to create %s: %w", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return "", "", fmt.Errorf("unable to write %s: %w", path, err)
	}

	return path, status, nil
}
