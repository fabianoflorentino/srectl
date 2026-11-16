package cli

import (
	"github.com/spf13/cobra"
)

func newCredsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "creds",
		Short: "Rotate the Argo CD admin password",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			password, err := cmd.Flags().GetString("password")
			if err != nil {
				return err
			}
			newPassword, err := newArgocd(cmd).RotatePassword(cmd.Context(), password)
			if err != nil {
				return err
			}
			return writef(cmd.OutOrStdout(), "Argo CD admin password: %s\n", newPassword)
		},
	}
	cmd.Flags().String("password", "", "new admin password (default: generate a random one)")
	return cmd
}
