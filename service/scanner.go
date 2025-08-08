package service

import (
	"log"
	"net"
	"network-scanner/model"
	"network-scanner/repository"
	"time"

	"github.com/j-keck/arping"
	"github.com/tatsushid/go-fastping"
)

type ScannerService struct {
	repo *repository.InMemoryRepository
}

func NewScannerService(r *repository.InMemoryRepository) *ScannerService {
	return &ScannerService{repo: r}
}

func (s *ScannerService) StartScan(ipRange string) {
	go func() {
		ips, err := getIPList(ipRange)
		if err != nil {
			log.Printf("Invalid CIDR: %s", ipRange)
			return
		}
		for _, ip := range ips {
			reachable := ping(ip)

			hostname := ""
			mac := ""

			if reachable {
				hostname = resolveHostname(ip)
				mac = resolveMAC(ip)
			}

			device := model.Device{
				IPAddress:  ip,
				MACAddress: mac,
				Hostname:   hostname,
				IsOnline:   reachable,
				LastSeen:   time.Now(),
			}
			s.repo.Save(device)
		}
	}()
}

func (s *ScannerService) GetDevices() []model.Device {
	return s.repo.GetAll()
}

func ping(ip string) bool {
	p := fastping.NewPinger()
	addr, err := net.ResolveIPAddr("ip4:icmp", ip)
	if err != nil {
		return false
	}

	p.AddIPAddr(addr)

	reachable := false
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		reachable = true
	}

	err = p.Run()
	if err != nil {
		return false
	}

	return reachable
}

func getIPList(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	if len(ips) > 2 {
		ips = ips[1 : len(ips)-1]
	}
	return ips, nil
}

func resolveHostname(ip string) string {
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return ""
	}
	return names[0]
}

func resolveMAC(ip string) string {
	ipAddr := net.ParseIP(ip)
	mac, _, err := arping.Ping(ipAddr)
	if err != nil {
		return ""
	}
	return mac.String()
}

func (s *ScannerService) Clear() {
	s.repo.Clear()
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
