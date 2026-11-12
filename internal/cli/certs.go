package cli

import (
	"github.com/spf13/cobra"
)

func newCertsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "certs",
		Short: "Trigger the certificate-validate scan",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return notImplemented("Phase 6")
		},
	}
}
