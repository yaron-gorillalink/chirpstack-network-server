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

// RequestRXParamSetup modifies the RX1 data-rate offset, RX2 frequency and
// RX2 data-rate.
func RequestRXParamSetup(rx1DROffset int, rx2Frequency uint32, rx2DR int) storage.MACCommandBlock {
	return storage.MACCommandBlock{
		CID: lorawan.RXParamSetupReq,
		MACCommands: []lorawan.MACCommand{
			{
				CID: lorawan.RXParamSetupReq,
				Payload: &lorawan.RXParamSetupReqPayload{
					Frequency: rx2Frequency,
					DLSettings: lorawan.DLSettings{
						RX2DataRate: uint8(rx2DR),
						RX1DROffset: uint8(rx1DROffset),
					},
				},
			},
		},
	}
}

func handleRXParamSetupAns(ctx context.Context, ds *storage.DeviceSession, block storage.MACCommandBlock, pendingBlock *storage.MACCommandBlock) ([]storage.MACCommandBlock, error) {
	if len(block.MACCommands) != 1 {
		return nil, fmt.Errorf("exactly one mac-command expected, got: %d", len(block.MACCommands))
	}

	if pendingBlock == nil || len(pendingBlock.MACCommands) == 0 {
		return nil, errors.New("expected pending mac-command")
	}
	req := pendingBlock.MACCommands[0].Payload.(*lorawan.RXParamSetupReqPayload)

	pl, ok := block.MACCommands[0].Payload.(*lorawan.RXParamSetupAnsPayload)
	if !ok {
		return nil, fmt.Errorf("expected *lorawan.RXParamSetupAnsPayload, got %T", block.MACCommands[0].Payload)
	}

	if !pl.ChannelACK || !pl.RX1DROffsetACK || !pl.RX2DataRateACK {
		// increase the error counter
		ds.MACCommandErrorCount[lorawan.RXParamSetupAns]++

		log.WithFields(log.Fields{
			"dev_eui":           ds.DevEUI,
			"channel_ack":       pl.ChannelACK,
			"rx1_dr_offset_ack": pl.RX1DROffsetACK,
			"rx2_dr_ack":        pl.RX2DataRateACK,
			"ctx_id":            ctx.Value(logging.ContextIDKey),
		}).Warning("rx_param_setup not acknowledged")
		return nil, nil
	}

	// reset the error counter
	delete(ds.MACCommandErrorCount, lorawan.RXParamSetupAns)

	ds.RX2Frequency = req.Frequency
	ds.RX2DR = req.DLSettings.RX2DataRate
	ds.RX1DROffset = req.DLSettings.RX1DROffset

	log.WithFields(log.Fields{
		"dev_eui":       ds.DevEUI,
		"rx2_frequency": req.Frequency,
		"rx2_dr":        req.DLSettings.RX2DataRate,
		"rx1_dr_offset": req.DLSettings.RX1DROffset,
		"ctx_id":        ctx.Value(logging.ContextIDKey),
	}).Info("rx_param_setup request acknowledged")

	return nil, nil
}
