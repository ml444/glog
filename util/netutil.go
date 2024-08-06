package util

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

func IsLocalIpV4(ip net.IP) bool {
	local := false
	if ip == nil {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil {
		switch true {
		case ip4[0] == 10:
			local = true
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			local = true
		case ip4[0] == 192 && ip4[1] == 168:
			local = true
		}
	}
	return local
}
func GetAllLocalIp() ([]string, error) {
	addressList, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	var ipList []string
	for _, address := range addressList {
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if IsLocalIpV4(ipNet.IP) {
				ipList = append(ipList, ipNet.IP.String())
			}
		}
	}
	return ipList, nil
}

func GetFirstLocalIp() (string, error) {
	ipList, err := GetAllLocalIp()
	if err != nil {
		//log.Errorf("Err: %v\n", err)
		return "", err
	}
	if len(ipList) > 0 {
		ip := ipList[0]
		return ip, nil
	}
	return "", err
}

func GetIpV4ByIFace(name string) (string, error) {
	ifList, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, v := range ifList {
		if v.Name == name {
			addressList, err := v.Addrs()
			if err != nil {
				return "", err
			}
			for _, address := range addressList {
				if ipNet, ok := address.(*net.IPNet); ok {
					if ip4 := ipNet.IP.To4(); ip4 != nil {
						return ipNet.IP.String(), nil
					}
				}
			}
		}
	}
	return "", errors.New(fmt.Sprintf("not found face %s", name))
}
func GetAllIFaceName() ([]string, error) {
	ifList, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var nameList []string
	for _, v := range ifList {
		nameList = append(nameList, v.Name)
	}
	return nameList, nil
}
func IpStr2Int(ip string) uint32 {
	ips := net.ParseIP(ip)
	if len(ips) == 16 {
		return binary.BigEndian.Uint32(ips[12:16])
	} else if len(ips) == 4 {
		return binary.BigEndian.Uint32(ips)
	}
	return 0
}
func IpInt2Str(ip uint32) string {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, ip)
	if err != nil {
		return ""
	}
	b := buf.Bytes()
	return fmt.Sprintf("%d.%d.%d.%d", b[0], b[1], b[2], b[3])
}
