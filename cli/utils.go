package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/0xfe/microstellar"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func showSuccess(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
}

func showError(fields logrus.Fields, msg string, args ...interface{}) {
	logrus.WithFields(fields).Errorf(msg, args...)
}

func (cli *CLI) validateAddressOrSeed(fields logrus.Fields, addressOrSeed string, keyType string) (string, error) {
	var err error
	key := addressOrSeed

	if !microstellar.ValidAddressOrSeed(addressOrSeed) {
		addressOrSeed, err = cli.GetAccountOrSeed(addressOrSeed, keyType)
		if err != nil {
			logrus.WithFields(fields).Debugf("invalid address, seed, or account name: %s", key)
			return "", err
		}
	}

	return addressOrSeed, nil
}

func (cli *CLI) genTxOptions(cmd *cobra.Command, logFields logrus.Fields) (*microstellar.TxOptions, error) {
	opts := microstellar.Opts()

	if memotext, err := cmd.Flags().GetString("memotext"); err == nil && memotext != "" {
		opts = opts.WithMemoText(memotext)
	}

	if memoid, err := cmd.Flags().GetString("memoid"); err == nil && memoid != "" {
		id, err := strconv.ParseUint(memoid, 10, 64)
		if err != nil {
			logrus.WithFields(logFields).Debugf("error parsing memoid: %v", err)
			return nil, errors.Errorf("bad memoid: %s", memoid)
		}
		opts = opts.WithMemoID(id)
	}

	return opts, nil
}

// GetAccount returns the account address or seed for "name". Set keyType
// to "address" or "seed" to specify the return value.
func (cli *CLI) GetAccount(name, keyType string) (string, error) {
	if keyType != "address" && keyType != "seed" {
		return name, errors.Errorf("invalid key type: %s", keyType)
	}

	key := fmt.Sprintf("account:%s:%s", name, keyType)

	code, err := cli.GetVar(key)

	if err != nil {
		return name, err
	}

	return code, nil
}

// GetAccountOrSeed returns the account address or seed for "name". It prefers
// keyType ("address" or "seed")
func (cli *CLI) GetAccountOrSeed(name, keyType string) (string, error) {
	code, err := cli.GetAccount(name, keyType)

	if err != nil {
		if keyType == "address" {
			keyType = "seed"
		} else {
			keyType = "address"
		}

		code, err = cli.GetAccount(name, keyType)
	}

	return code, err
}

// GetAsset returns the asset with the given name
func (cli *CLI) GetAsset(name string) (*microstellar.Asset, error) {
	if name == "" {
		return microstellar.NativeAsset, nil
	}

	readField := func(field string) (string, error) {
		key := fmt.Sprintf("asset:%s:%s", name, field)
		val, err := cli.GetVar(key)
		if err != nil {
			return "", err
		}

		return val, nil
	}

	code, err1 := readField("code")
	issuer, err2 := readField("issuer")
	assetType, err3 := readField("type")

	if err1 != nil || err2 != nil || err3 != nil {
		return nil, errors.Errorf("could not read asset: %v, %v, %v", err1, err2, err3)
	}

	var asset *microstellar.Asset

	if assetType == string(microstellar.Credit4Type) {
		asset = microstellar.NewAsset(code, issuer, microstellar.Credit4Type)
	} else if assetType == string(microstellar.Credit12Type) {
		asset = microstellar.NewAsset(code, issuer, microstellar.Credit12Type)
	} else {
		asset = microstellar.NativeAsset
	}

	logrus.Debugf("got asset: %+v", asset)
	return asset, nil
}
