package maccommand

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/chirpstack-network-server/v3/internal/logging"
	"github.com/brocaar/chirpstack-network-server/v3/internal/storage"
	"github.com/brocaar/lorawan"
)

// RequestPingSlotChannel modifies the frequency and / or the data-rate
// on which the end-device expects the downlink pings (class-b).
func RequestPingSlotChannel(devEUI lorawan.EUI64, dr int, freq uint32) storage.MACCommandBlock {
	return storage.MACCommandBlock{
		CID: lorawan.PingSlotChannelReq,
		MACCommands: []lorawan.MACCommand{
			{
				CID: lorawan.PingSlotChannelReq,
				Payload: &lorawan.PingSlotChannelReqPayload{
					Frequency: freq,
					DR:        uint8(dr),
				},
			},
		},
	}
}

func handlePingSlotChannelAns(ctx context.Context, ds *storage.DeviceSession, block storage.MACCommandBlock, pendingBlock *storage.MACCommandBlock) ([]storage.MACCommandBlock, error) {
	if len(block.MACCommands) != 1 {
		return nil, fmt.Errorf("exactly one mac-command expected, got: %d", len(block.MACCommands))
	}

	if pendingBlock == nil || len(pendingBlock.MACCommands) == 0 {
		return nil, errors.New("expected pending mac-command")
	}
	req := pendingBlock.MACCommands[0].Payload.(*lorawan.PingSlotChannelReqPayload)

	pl, ok := block.MACCommands[0].Payload.(*lorawan.PingSlotChannelAnsPayload)
	if !ok {
		return nil, fmt.Errorf("expected *lorawan.PingSlotChannelAnsPayload, got %T", block.MACCommands[0].Payload)
	}

	if !pl.ChannelFrequencyOK || !pl.DataRateOK {
		// increase the error counter
		ds.MACCommandErrorCount[lorawan.PingSlotChannelAns]++

		log.WithFields(log.Fields{
			"dev_eui":              ds.DevEUI,
			"channel_frequency_ok": pl.ChannelFrequencyOK,
			"data_rate_ok":         pl.DataRateOK,
			"ctx_id":               ctx.Value(logging.ContextIDKey),
		}).Warning("ping_slot_channel request not acknowledged")
		return nil, nil
	}

	// reset the error counter
	delete(ds.MACCommandErrorCount, lorawan.PingSlotChannelAns)

	ds.PingSlotDR = int(req.DR)
	ds.PingSlotFrequency = req.Frequency

	log.WithFields(log.Fields{
		"dev_eui":           ds.DevEUI,
		"channel_frequency": ds.PingSlotFrequency,
		"data_rate":         ds.PingSlotDR,
		"ctx_id":            ctx.Value(logging.ContextIDKey),
	}).Info("ping_slot_channel request acknowledged")

	return nil, nil
}
