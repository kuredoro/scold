package util_test

import (
	"testing"

	"github.com/kuredoro/scold/util"
)

func TestSpyWaitGroup(t *testing.T) {
	t.Run("normal usage, 1 goroutine", func(t *testing.T) {
		var wg util.SpyWaitGroup
		wg.Add(1)
		go func() {
			wg.Done()
		}()

		wg.Wait()

		util.AssertSpyWaitGroupNormalUsage(t, &wg, 1)
	})

	t.Run("normal usage, 2 goroutines", func(t *testing.T) {
		var wg util.SpyWaitGroup
		wg.Add(2)
		go func() {
			wg.Done()
		}()

		go func() {
			wg.Done()
		}()

		wg.Wait()

		util.AssertSpyWaitGroupNormalUsage(t, &wg, 2)
	})
}
