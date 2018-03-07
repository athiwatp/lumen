package cli

import (
	"github.com/0xfe/microstellar"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (cli *CLI) getPayCmd() *cobra.Command {
	payCmd := &cobra.Command{
		Use:   "pay [amount] [asset] --from [source] --to [target]",
		Short: "send [amount] of [asset] from [source] to [target]",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fields := logrus.Fields{"cmd": "pay"}
			amount := args[0]
			assetName := ""
			if len(args) > 1 {
				assetName = args[1]
			}

			asset, err := cli.GetAsset(assetName)
			if err != nil {
				logrus.WithFields(fields).Debugf("could not get asset %s: %v", assetName, err)
				cli.error(fields, "bad asset: %s", assetName)
				return
			}

			to, _ := cmd.Flags().GetString("to")
			from, _ := cmd.Flags().GetString("from")

			source, err := cli.validateAddressOrSeed(fields, from, "seed")
			target, err := cli.validateAddressOrSeed(fields, to, "address")

			if err != nil {
				cli.error(fields, "bad source or target address")
				return
			}

			opts, err := cli.genTxOptions(cmd, fields)

			if err != nil {
				cli.error(fields, "can't generate payment: %v", err)
				return
			}

			fund, err := cmd.Flags().GetBool("fund")

			if fund {
				logrus.WithFields(fields).Debugf("initial fund from %s to %s, opts: %+v", source, target, opts)
				err = cli.ms.FundAccount(source, target, amount, opts)
			} else {
				logrus.WithFields(fields).Debugf("paying %s %s/%s from %s to %s, opts: %+v", amount, asset.Code, asset.Issuer, source, target, opts)
				err = cli.ms.Pay(source, target, amount, asset, opts)
			}

			if err != nil {
				cli.error(fields, "payment failed: %v", microstellar.ErrorString(err))
				return
			}
		},
	}

	payCmd.Flags().String("from", "", "source account seed or name")
	payCmd.Flags().String("to", "", "target account address or name")
	payCmd.Flags().Bool("fund", false, "fund a new account")
	payCmd.MarkFlagRequired("from")
	payCmd.MarkFlagRequired("to")

	// Transaction envelope options
	payCmd.Flags().String("memotext", "", "memo text")
	payCmd.Flags().String("memoid", "", "memo ID")

	return payCmd
}
