package network

import (
	"net"
	"slices"
)

func Interfaces() Cards {
	faces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	var results []*Card
	for _, iface := range faces {
		// 跳过未启用或环回接口
		if iface.Flags&net.FlagUp == 0 ||
			iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, _ := iface.Addrs()
		info := &Card{
			Index: iface.Index,
			Name:  iface.Name,
			MTU:   iface.MTU,
			MAC:   iface.HardwareAddr.String(),
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}

			// 过滤无效地址
			if ip == nil ||
				ip.IsLoopback() ||
				ip.IsMulticast() ||
				ip.IsUnspecified() {
				continue
			}

			if ip4 := ip.To4(); ip4 != nil {
				info.IPv4 = append(info.IPv4, ip4.String())
			} else if ip.To16() != nil {
				// 排除 IPv6 链路本地地址（fe80::/10），如不需要可移除此条件
				if ip.IsLinkLocalUnicast() {
					continue
				}
				info.IPv6 = append(info.IPv6, ip.String())
			}
		}

		// 没有有效地址就跳过
		if len(info.IPv4) == 0 && len(info.IPv6) == 0 {
			continue
		}
		results = append(results, info)
	}

	return results
}

type Card struct {
	Name  string   `json:"name"`
	Index int      `json:"index"`
	MTU   int      `json:"mtu"`
	IPv4  []string `json:"ipv4"`
	IPv6  []string `json:"ipv6"`
	MAC   string   `json:"mac"`
}

func (c Card) equal(v *Card) bool {
	if c.Name != v.Name ||
		c.Index != v.Index ||
		c.MTU != v.MTU ||
		c.MAC != v.MAC {
		return false
	}

	return slices.Equal(c.IPv4, v.IPv4) &&
		slices.Equal(c.IPv6, v.IPv6)
}

type Cards []*Card

func (cs Cards) Equal(vs Cards) bool {
	if len(cs) != len(vs) {
		return false
	}

	for i, c := range cs {
		if !c.equal(vs[i]) {
			return false
		}
	}

	return true
}
