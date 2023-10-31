package partida

import "net"

// fetch the first IPv4 address from an FQDN
func GetFQDNIPv4Address(fqdn string) (addr string) {
	ips, _ := net.LookupIP(fqdn)
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			addr = ipv4.String()
		}
	}

	return
}

// fetch the reverse dns address from an IP
func GetIPAddressRDNS(addr string) (rdns string) {
	addresses, _ := net.LookupAddr(addr)
	return addresses[0]
}
