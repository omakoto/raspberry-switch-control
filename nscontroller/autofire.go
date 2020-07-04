package nscontroller

import (
	"github.com/omakoto/raspberry-switch-control/nscontroller/utils"
	"time"
)

type AutoFirer struct {
	syncer *utils.Synchronized

	next Consumer

	// intervals is the autofire interval for each button
	intervals []time.Duration
}

func NewAutoFirer(next Consumer) *AutoFirer {
	return &AutoFirer{utils.NewSynchronized(), next, make([]time.Duration, NumActionButton)}
}

func (af *AutoFirer) SetFireInterval(a Action, interval time.Duration) {
	af.syncer.Run(func() {
		af.intervals[a] = interval
	})
}

func (af *AutoFirer) Consume(ev *Event) {
	af.syncer.Run(func() {
		af.next(ev)
	})
}
