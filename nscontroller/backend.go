package nscontroller

import "io"

type BackendConsumer struct {
	out io.WriteCloser
	ch  chan Event
}

var _ Consumer = (*BackendConsumer)(nil)
var _ Worker = (*BackendConsumer)(nil)

func NewBackendConsumer(out io.WriteCloser) (*BackendConsumer, error) {
	return &BackendConsumer{out, make(chan Event)}, nil
}

func (b *BackendConsumer) Close() error {
	close(b.ch)
	return b.out.Close()
}

func (b *BackendConsumer) Intake() <-chan Event {
	return b.ch
}

func (b *BackendConsumer) Run() {
	go func() {
		for ;; {
			ev := <- b.ch

			switch ev.Action {

			}
		}
	}()
}
