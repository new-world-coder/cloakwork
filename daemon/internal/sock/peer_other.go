//go:build !darwin

package sock

import (
	"fmt"
	"net"
)

func peerUID(conn *net.UnixConn) (int, error) {
	return 0, fmt.Errorf("peer credentials are only implemented on macOS")
}
