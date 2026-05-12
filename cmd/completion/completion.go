package completion

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"openhue-cli/openhue"
)

type shellProvider func() string

// NewCmdCompletion creates the command tree for generating shell completions.
func NewCmdCompletion(io openhue.IOStreams) *cobra.Command {
	return newCmdCompletion(io.Out, currentShell)
}

func newCmdCompletion(out io.Writer, getShell shellProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "completion",
		GroupID:       "config",
		Short:         "Generate shell completion scripts",
		SilenceErrors: true,
		SilenceUsage:  true,
		Long: `
Generate shell completion scripts for openhue.

Use one of the shell-specific subcommands to print a completion script, or use
the setup subcommand to print the command for enabling completions in your
current shell session.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return printSetupCommand(out, getShell())
		},
	}

	cmd.AddCommand(newCmdCompletionBash(out))
	cmd.AddCommand(newCmdCompletionZsh(out))
	cmd.AddCommand(newCmdCompletionFish(out))
	cmd.AddCommand(newCmdCompletionSetup(out, getShell))

	return cmd
}

func newCmdCompletionBash(out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "bash",
		Short: "Generate the bash completion script",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenBashCompletionV2(out, true)
		},
	}
}

func newCmdCompletionZsh(out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "zsh",
		Short: "Generate the zsh completion script",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenZshCompletion(out)
		},
	}
}

func newCmdCompletionFish(out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "fish",
		Short: "Generate the fish completion script",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenFishCompletion(out, true)
		},
	}
}

func newCmdCompletionSetup(out io.Writer, getShell shellProvider) *cobra.Command {
	return &cobra.Command{
		Use:           "setup",
		Short:         "Print the command to enable completions for your shell",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return printSetupCommand(out, getShell())
		},
	}
}

func printSetupCommand(out io.Writer, shell string) error {
	command, err := setupCommandForShell(shell)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "Run this command to enable openhue completions for your current shell session:\n\n  %s\n", command)
	return err
}

func setupCommandForShell(shell string) (string, error) {
	switch filepath.Base(shell) {
	case "bash":
		return "source <(openhue completion bash)", nil
	case "zsh":
		return "source <(openhue completion zsh)", nil
	case "fish":
		return "openhue completion fish | source", nil
	default:
		return "", fmt.Errorf("unsupported shell %q; supported shells are bash, zsh, and fish", shell)
	}
}

func currentShell() string {
	return os.Getenv("SHELL")
}
