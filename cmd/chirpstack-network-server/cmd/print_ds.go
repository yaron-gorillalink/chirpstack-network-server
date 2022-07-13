package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/brocaar/chirpstack-network-server/v3/internal/config"
	"github.com/brocaar/chirpstack-network-server/v3/internal/storage"
	"github.com/brocaar/lorawan"
)

var printDSCmd = &cobra.Command{
	Use:     "print-ds",
	Short:   "Print the device-session as JSON (for debugging)",
	Example: `chirpstack-network-server print-ds 0102030405060708`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatalf("hex encoded DevEUI must be given as an argument")
		}

		if err := storage.Setup(config.C); err != nil {
			log.Fatal(err)
		}

		var devEUI lorawan.EUI64
		if err := devEUI.UnmarshalText([]byte(args[0])); err != nil {
			log.WithError(err).Fatal("decode DevEUI error")
		}

		ds, err := storage.GetDeviceSession(context.Background(), devEUI)
		if err != nil {
			log.WithError(err).Fatal("get device-session error")
		}

		b, err := json.MarshalIndent(ds, "", "    ")
		if err != nil {
			log.WithError(err).Fatal("json marshal error")
		}

		fmt.Println(string(b))
	},
}
