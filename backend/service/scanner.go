package service

import (
	"context"
	"log"
	"net"
	"network-scanner/model"
	"network-scanner/repository"
	"sync"
	"time"

	"github.com/j-keck/arping"
	"github.com/tatsushid/go-fastping"
)

type ScannerService struct {
	repo   *repository.InMemoryRepository
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewScannerService(r *repository.InMemoryRepository) *ScannerService {
	return &ScannerService{repo: r}
}

func (s *ScannerService) StartScan(ipRange string) {
	if s.cancel != nil {
		s.cancel()
		s.wg.Wait()
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		ips, err := getIPList(ipRange)
		if err != nil {
			log.Printf("Invalid CIDR: %s", ipRange)
			return
		}
		for _, ip := range ips {
			select {
			case <-ctx.Done():
				log.Println("Scan cancelled")
				return
			default:
				reachable := ping(ip)

				hostname := ""
				mac := ""
				existing := s.repo.FindByIP(ip)
				lastSeen := time.Time{}

				if reachable {
					hostname = resolveHostname(ip)
					mac = resolveMAC(ip)
					lastSeen = time.Now()
				} else if existing != nil {
					lastSeen = existing.LastSeen
				}

				device := model.Device{
					IPAddress:  ip,
					MACAddress: mac,
					Hostname:   hostname,
					IsOnline:   reachable,
					LastSeen:   lastSeen,
				}
				s.repo.Save(device)
			}
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
	if ipAddr.IsLoopback() {
		return ""
	}
	mac, _, err := arping.Ping(ipAddr)
	if err != nil {
		return ""
	}
	return mac.String()
}

func (s *ScannerService) Clear() {
	if s.cancel != nil {
		s.cancel()
		s.wg.Wait()
	}
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
