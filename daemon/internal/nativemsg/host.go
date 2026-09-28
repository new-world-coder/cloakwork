// Package nativemsg is the Chrome native messaging stdio host.
// Chrome authenticates the caller by launching this process itself and
// checking allowed_origins in the host manifest. Do not write logs to stdout.
package nativemsg

import (
	"io"

	"github.com/cloakwork/cloakwork/daemon/internal/protocol"
)

// Host serves length-prefixed JSON on the stdio pipes Chrome provides.
func Host(r io.Reader, w io.Writer, handle func([]byte) []byte) error {
	return protocol.Serve(r, w, handle)
}
