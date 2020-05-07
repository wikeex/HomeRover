package base

import (
	"net"
	"testing"
)

func TestAllocatePort(t *testing.T)  {
	var conn *net.UDPConn
	defer func() {
		err := conn.Close()
		if err != nil {
			t.Error(err)
		}
	}()

	addr, err := allocatePort(&conn)
	if err != nil {
		t.Error(err)
	}
	t.Log(addr.String())
	t.Log(conn.LocalAddr().String())
}
