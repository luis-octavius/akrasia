//go:build !unix

package web

// Lock does nothing where flock isn't available: `restore` can't tell
// whether `serve` is running there, so stopping it first is up to the user.
func Lock(string) (func(), error) { return func() {}, nil }

// Running always reports false where Lock does nothing.
func Running(string) bool { return false }
