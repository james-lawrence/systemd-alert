package influxdb

import (
	"bytes"
	"net"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

func newUnixSocket(address string) unixClient {
	return unixClient{
		Address: address,
	}
}

// unixClient - client for connecting to influxdb through unix sockets.
type unixClient struct {
	Address string
}

// Ping checks that status of cluster, and will always return 0 time and no
// error for UDP clients.
func (t unixClient) Ping(timeout time.Duration) (time.Duration, string, error) {
	return time.Duration(0), "", nil
}

// Write takes a BatchPoints object and writes all Points to InfluxDB.
func (t unixClient) Write(bp client.BatchPoints) error {
	var (
		err  error
		conn net.Conn
		b    bytes.Buffer
	)

	for _, p := range bp.Points() {
		if _, err = b.WriteString(p.PrecisionString(bp.Precision())); err != nil {
			return err
		}

		if err = b.WriteByte('\n'); err != nil {
			return err
		}
	}

	if conn, err = net.Dial("unix", t.Address); err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(b.Bytes())
	return err
}
