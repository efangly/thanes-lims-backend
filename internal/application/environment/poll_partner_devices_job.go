package environment

import (
	"context"
	"log"

	portenvironment "github.com/efangly/thanes-lims-backend/internal/ports/environment"
)

// PollPartnerDevicesJob is the scheduled counterpart that keeps every
// Active Partner Device's cached snapshot fresh - run on a fixed interval
// (PARTNER_API_POLL_INTERVAL) from cmd/api/main.go, mirroring
// applicationpurchaseorder.AutoReorderJob.
type PollPartnerDevicesJob struct {
	devices portenvironment.PartnerDeviceRepository
	poll    *PollPartnerDeviceUseCase
}

func NewPollPartnerDevicesJob(devices portenvironment.PartnerDeviceRepository, poll *PollPartnerDeviceUseCase) *PollPartnerDevicesJob {
	return &PollPartnerDevicesJob{devices: devices, poll: poll}
}

// Run polls every Active Partner Device once. One device's failure is
// logged and skipped rather than aborting the rest of the batch.
func (j *PollPartnerDevicesJob) Run(ctx context.Context) {
	devices, err := j.devices.List(ctx)
	if err != nil {
		log.Printf("partnerdevice poll: list devices: %v", err)
		return
	}
	for _, d := range devices {
		if !d.Active {
			continue
		}
		if _, err := j.poll.Execute(ctx, d); err != nil {
			log.Printf("partnerdevice poll: %s: %v", d.Serial, err)
		}
	}
}
