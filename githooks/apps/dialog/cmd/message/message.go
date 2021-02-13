package message

import (
	"context"
	dcm "gabyx/githooks/apps/dialog/cmd/common"
	"gabyx/githooks/apps/dialog/gui"
	res "gabyx/githooks/apps/dialog/result"
	set "gabyx/githooks/apps/dialog/settings"
	ccm "gabyx/githooks/cmd/common"
	"time"

	"github.com/spf13/cobra"
)

func handleResult(ctx *dcm.CmdContext, res *res.Message, err error) error {
	return dcm.HandleGeneralResult(ctx, &res.General, err, nil, nil)
}

func NewCmd(ctx *dcm.CmdContext) *cobra.Command {

	settings := set.Message{}
	var timeout uint

	cmd := &cobra.Command{
		Use:   "message",
		Short: "Shows a message  dialog.",
		Long: `Shows a message dialog similar to 'zenity'.

# Exit Codes:

- '0' : User pressed 'Ok'.
- '1' : User pressed 'Cancel' or closed the dialog.
- '2' : The user pressed an extra button.
		The output contains the index of that button.`,

		Run: func(cmd *cobra.Command, args []string) {

			var cancel func()
			var cont context.Context

			if timeout > 0 {
				cont, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
				defer cancel()
			}

			res, err := gui.ShowMessage(cont, &settings)
			err = handleResult(ctx, &res, err)
			ctx.Log.AssertNoErrorPanic(err, "Dialog failed")
		}}

	cmd.Flags().UintVar(&timeout, "timeout", 0, "Timeout for the dialog")

	dcm.AddFlagsMessage(cmd, &settings)
	ccm.SetCommandDefaults(ctx.Log, cmd)

	return cmd
}
