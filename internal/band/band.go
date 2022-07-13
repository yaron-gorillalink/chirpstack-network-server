package band

import (
	"github.com/pkg/errors"

	"github.com/brocaar/chirpstack-network-server/v3/internal/config"
	"github.com/brocaar/lorawan"
	loraband "github.com/brocaar/lorawan/band"
)

var band loraband.Band

// Setup sets up the band with the given configuration.
func Setup(c config.Config) error {
	dwellTime := lorawan.DwellTimeNoLimit
	if c.NetworkServer.Band.DownlinkDwellTime400ms {
		dwellTime = lorawan.DwellTime400ms
	}
	bandConfig, err := loraband.GetConfig(c.NetworkServer.Band.Name, c.NetworkServer.Band.RepeaterCompatible, dwellTime)
	if err != nil {
		return errors.Wrap(err, "get band config error")
	}
	for _, ec := range c.NetworkServer.NetworkSettings.ExtraChannels {
		if err := bandConfig.AddChannel(ec.Frequency, ec.MinDR, ec.MaxDR); err != nil {
			return errors.Wrap(err, "add channel error")
		}
	}
	band = bandConfig
	return nil
}

// Band returns the configured band.
func Band() loraband.Band {
	return band
}
