package agent

import "net"

func LocalOutboundIP(remoteAddr string) string {
	conn, err := net.Dial("udp", remoteAddr)
	if err != nil {
		return ""
	}
	defer conn.Close()

	if udpAddr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		return udpAddr.IP.String()
	}
	return ""
}
