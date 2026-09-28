//go:build darwin

package sock

import (
	"net"

	"golang.org/x/sys/unix"
)

func peerUID(conn *net.UnixConn) (int, error) {
	raw, err := conn.SyscallConn()
	if err != nil {
		return 0, err
	}
	var (
		uid  int
		gerr error
	)
	// LOCAL_PEERCRED is the macOS peer-uid check (the same fact getpeereid reports).
	err = raw.Control(func(fd uintptr) {
		cred, err := unix.GetsockoptXucred(int(fd), unix.SOL_LOCAL, unix.LOCAL_PEERCRED)
		if err != nil {
			gerr = err
			return
		}
		uid = int(cred.Uid)
	})
	if err != nil {
		return 0, err
	}
	return uid, gerr
}
