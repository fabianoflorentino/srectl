// Package cli builds the srectl Cobra command tree. Each command is a thin
// task-runner step that wires the internal packages together.
package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/fabianoflorentino/srectl/internal/config"
)

type configContextKey struct{}

func configFrom(cmd *cobra.Command) *config.Config {
	if v := cmd.Context().Value(configContextKey{}); v != nil {
		return v.(*config.Config)
	}
	return nil
}

func notImplemented(phase string) error {
	return fmt.Errorf("not implemented yet (%s)", phase)
}

// writef writes formatted output to w, propagating any write error.
func writef(w io.Writer, format string, a ...any) error {
	_, err := fmt.Fprintf(w, format, a...)
	return err
}

// stubCommand builds a leaf command that reports the feature as not yet
// implemented. Stubs keep the full command surface discoverable via --help.
func stubCommand(use, short, phase string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return notImplemented(phase)
		},
	}
}

// NewRootCommand builds the full srectl command tree.
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "srectl",
		Short: "A local SRE/DevOps/Platform Engineering training environment on kind",
		Long: "srectl is a task runner (like Make) that stands up and operates a full\n" +
			"platform on a local kind cluster: GitOps (Argo CD), observability (LGTM),\n" +
			"security, chaos engineering and demo workloads.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			v := config.NewViper()
			if f := cmd.Flag("config").Value.String(); f != "" {
				v.SetConfigFile(f)
			}
			cfg, err := config.Load(v)
			if err != nil {
				return err
			}
			cmd.SetContext(context.WithValue(cmd.Context(), configContextKey{}, cfg))
			return nil
		},
	}

	root.PersistentFlags().String("config", "", "path to config file (default: config/srectl.yaml)")

	root.AddCommand(
		newVersionCommand(),
		newUpCommand(),
		newDownCommand(),
		newStatusCommand(),
		newClusterCommand(),
		newBootstrapCommand(),
		newSyncCommand(),
		newAppsCommand(),
		newDemoCommand(),
		newTraceCommand(),
		newCertsCommand(),
		newOpenCommand(),
		newLogsCommand(),
	)

	return root
}

// Execute runs the root command and returns its error.
func Execute() error {
	return NewRootCommand().Execute()
}
