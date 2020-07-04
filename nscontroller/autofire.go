package nscontroller

import (
	"github.com/omakoto/go-common/src/common"
	"github.com/omakoto/raspberry-switch-control/nscontroller/utils"
	"github.com/pborman/getopt/v2"
	"time"
)

var (
	tickInterval = getopt.IntLong("tick", 't', 10, "Tick interval in milliseconds")
)

type AutoFirer struct {
	syncer *utils.Synchronized

	next Consumer

	// intervals is the autofire interval for each button
	intervals []time.Duration

	lastTimestamp []time.Duration
	nextTimestamp []time.Duration

	ticker *time.Ticker
	stop   chan bool
}

var _ Worker = (*AutoFirer)(nil)

func NewAutoFirer(next Consumer) *AutoFirer {
	return &AutoFirer{
		utils.NewSynchronized(),
		next,
		make([]time.Duration, NumActionButton),
		make([]time.Duration, NumActionButton),
		make([]time.Duration, NumActionButton),
		nil,
		nil,
	}
}

func (af *AutoFirer) Run() {
	af.syncer.Run(func() {
		if af.ticker != nil {
			common.Fatal("AutoFirer already running")
			return
		}
		common.Debug("AutoFirer started")
		af.ticker = time.NewTicker(time.Duration(*tickInterval) * time.Millisecond)
		af.stop = make(chan bool)

		ticker := af.ticker
		stop := af.stop

		go func() {
		loop:
			for {
				select {
				case <-ticker.C:
					common.Debug("AutoFirer tick")
				case <-stop:
					break loop
				}
			}
			af.syncer.Run(func() {
				af.ticker = nil
				af.stop = nil
			})
			common.Debug("AutoFirer stopped")
		}()
	})
}

func (af *AutoFirer) Close() error {
	af.syncer.Run(func() {
		if af.ticker != nil {
			common.Debug("AutoFirer stopping")
			af.stop <- true
		} else {
			common.Debug("AutoFirer not running")
		}
	})
	return nil
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
