/*
Copyright © 2023 Glif LTD
*/
package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/briandowns/spinner"
	"github.com/ethereum/go-ethereum/common"
	"github.com/glif-confidential/cli/fevm"
	"github.com/spf13/cobra"
)

// borrowCmd represents the borrow command
var withdrawCmd = &cobra.Command{
	Use:   "withdraw <amount>",
	Short: "Withdraw FIL from your Agent.",
	Long:  "",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		agentAddr, ownerKey, err := commonSetupOwnerCall()
		if err != nil {
			log.Fatal(err)
		}

		conn := fevm.Connection()

		var receiver common.Address
		if cmd.Flag("recipient") != nil && cmd.Flag("recipient").Changed {
			receiver = common.HexToAddress(cmd.Flag("recipient").Value.String())
		} else {
			// if no recipient is specified, use the agent's owner
			receiver, err = conn.AgentOwner(cmd.Context(), agentAddr)
			if err != nil {
				log.Fatal(err)
			}
		}

		if common.IsHexAddress(receiver.String()) {
			log.Fatal("Invalid withdraw address")
		}

		amount, err := parseFILAmount(args[0])
		if err != nil {
			log.Fatal(err)
		}

		s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
		s.Start()

		fmt.Printf("Withdrawing %s FIL from your Agent...", args[0])

		tx, err := conn.AgentWithdraw(cmd.Context(), agentAddr, receiver, amount, ownerKey)
		if err != nil {
			log.Fatal(err)
		}

		// transaction landed on chain or errored
		receipt, err := fevm.WaitReturnReceipt(tx.Hash())
		if err != nil {
			log.Fatal(err)
		}

		if receipt == nil {
			log.Fatal("Failed to get receipt")
		}

		if receipt.Status == 0 {
			log.Fatal("Transaction failed")
		}

		s.Stop()

		fmt.Printf("Successfully withdrew %s FIL", args[0])
	},
}

func init() {
	agentCmd.AddCommand(withdrawCmd)
	createCmd.Flags().String("recipient", "", "Receiver of the funds (agent's owner is chosen if not specified)")
}