package patrol

import (
	"encoding/json"
	"net"
)

func (self *Patrol) HandleUDPConnection(
	conn *net.UDPConn,
) error {
	// wrap this function in a for loop, externally to this package
	// if timeouts are necessary add them outside of that loop
	body := make([]byte, 2048)
	n, addr, err := conn.ReadFromUDP(body)
	if err != nil {
		// we failed to read
		return err
	}
	// read something
	request := &API_Request{}
	// unmarshal
	if err = json.Unmarshal(body[:n], request); err != nil ||
		!request.IsValid() {
		// failed to unmarshal
		// ignore error
		return nil
	}
	// read response
	response := self.api(api_endpoint_udp, request)
	// marshal response
	bs, _ := json.Marshal(response)
	// write response
	if _, err = conn.WriteToUDP(bs, addr); err != nil {
		return err
	}
	// done!
	return nil
}
