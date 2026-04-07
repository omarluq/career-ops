// Package closer provides helpers for handling deferred close errors.
package closer

import "io"

// Guard captures a named-return error pointer so that deferred
// Close calls can propagate close errors without losing the
// original error. Usage:
//
//	func foo() (err error) {
//	    g := closer.Guard{Err: &err}
//	    f, err := os.Open(...)
//	    if err != nil { return err }
//	    defer g.Close(f)
//	    ...
//	}
type Guard struct {
	Err *error
}

// Close calls c.Close() and, if the guarded error is still nil,
// replaces it with the close error.
func (g Guard) Close(c io.Closer) {
	if closeErr := c.Close(); closeErr != nil && *g.Err == nil {
		*g.Err = closeErr
	}
}
