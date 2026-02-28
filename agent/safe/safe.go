package safe

import (
	"opslense-pulse/shared"
)

func Go(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				shared.Error("goroutine panic recovered", "worker", name, "error", r)
			}
		}()
		fn()
	}()
}
