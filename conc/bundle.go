package conc

import (
	"context"
	"sync"
)

// A bundle is a set of goroutines that
// run together under a shared context.
//
type Bundle struct {
	wg              sync.WaitGroup
	bundleCtx       context.Context
	cancelBundleCtx func()
}

func NewBundle(parentContext context.Context) *Bundle {
	bundleCtx, cancelBundleCtx := context.WithCancel(parentContext)

	// XXX TODO I think with finalizers and bypassing the GC with cgo/unsafe/syscalls
	// we can make bundles cancel automatically when they are gc'd

	return &Bundle{
		wg:              sync.WaitGroup{},
		bundleCtx:       bundleCtx,
		cancelBundleCtx: cancelBundleCtx,
	}
}

func (b *Bundle) Go(task func(ctx context.Context)) {
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		task(b.bundleCtx)
	}()
}

// Cancel bundle goroutines without waiting for them to exit.
func (b *Bundle) Cancel() {
	b.cancelBundleCtx()
}

// Cancel bundle goroutines
// and wait until they are completed.
func (b *Bundle) Close() {
	b.Cancel()
	b.wg.Wait()
}
