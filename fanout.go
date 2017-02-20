package holster

import "sync"

// FanOut spawns a new go-routine each time `Run()` is called until `size` is reached,
// subsequent calls to `Run()` will block until previously `Run()` routines have completed.
// Allowing the user to control how many routines will run simultaneously. `Wait()` then
// collects any errors from the routines once they have all completed.
type FanOut struct {
	errChan chan error
	size    chan bool
	errs    []error
	routine sync.WaitGroup
}

func NewFanOut(size int) *FanOut {
	pool := FanOut{
		errChan: make(chan error, size),
		size:    make(chan bool, size),
		errs:    make([]error, 0),
	}
	pool.start()
	return &pool
}

func (p *FanOut) start() {
	p.routine.Add(1)
	go func() {
		for {
			select {
			case err, ok := <-p.errChan:
				if !ok {
					p.routine.Done()
					return
				}
				p.errs = append(p.errs, err)
			}
		}
	}()
}

// Run a new routine with an optional data value
func (p *FanOut) Run(callBack func(interface{}) error, data interface{}) {
	p.size <- true
	err := callBack(data)
	if err != nil {
		p.errChan <- err
	}
	<-p.size
}

// Wait for all the routines to complete and return any errors
func (p *FanOut) Wait() []error {
	// Wait for all the routines to complete
	for i := 0; i < cap(p.size); i++ {
		p.size <- true
	}
	// Close the err channel
	if p.errChan != nil {
		close(p.errChan)
	}

	// Wait until the error collector routine is complete
	p.routine.Wait()

	// If there are no errors
	if len(p.errs) == 0 {
		return nil
	}
	return p.errs
}