package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/mycicle/MyChain/blockchain/fs"
	"github.com/spf13/cobra"
)

const flagDataDir = "datadir"
const flagIP = "ip"
const flagPort = "port"

func main() {
	var tbbCmd = &cobra.Command{
		Use:   "tbb",
		Short: "The Blockchain Bar CLI",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	tbbCmd.AddCommand(versionCmd)
	tbbCmd.AddCommand(balancesCmd())
	tbbCmd.AddCommand(runCmd())
	tbbCmd.AddCommand(migrateCmd())

	err := tbbCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addDefaultRequiredFlags(cmd *cobra.Command) {
	cmd.Flags().String(
		flagDataDir,
		"",
		"Absolute path where all data will be / is stored",
	)
	cmd.MarkFlagRequired(flagDataDir)
}

func getDataDirFromCmd(cmd *cobra.Command) (string, error) {
	dataDir, err := cmd.Flags().GetString(flagDataDir)
	if err != nil {
		return fs.ExpandPath(dataDir), err
	}

	return fs.ExpandPath(dataDir), nil
}

func incorrectUsageErr() error {
	return errors.New("You used the cli incorrectly. Refer to tbb help for more information")
}
