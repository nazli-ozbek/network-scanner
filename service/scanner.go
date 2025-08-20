package service

import (
	"context"
	"net"
	"network-scanner/logger"
	"network-scanner/model"
	"network-scanner/repository"
	"sync"
	"time"

	"github.com/j-keck/arping"
	"github.com/tatsushid/go-fastping"
)

type ScannerService struct {
	repo        repository.DeviceRepository
	historyRepo repository.ScanHistoryRepository
	cancel      context.CancelFunc
	logger      logger.Logger
	wg          sync.WaitGroup
}

func NewScannerService(repo repository.DeviceRepository, history repository.ScanHistoryRepository, logger logger.Logger) *ScannerService {
	return &ScannerService{repo: repo, historyRepo: history, logger: logger}
}

func concurrentPing(ips []string, timeout time.Duration) map[string]bool {
	p := fastping.NewPinger()
	p.MaxRTT = timeout

	results := make(map[string]bool)
	var mu sync.Mutex

	for _, ip := range ips {
		addr, err := net.ResolveIPAddr("ip4:icmp", ip)
		if err == nil {
			p.AddIPAddr(addr)
		}
	}

	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		mu.Lock()
		results[addr.String()] = true
		mu.Unlock()
	}

	p.OnIdle = func() {}

	if err := p.Run(); err != nil {
		return results
	}

	for _, ip := range ips {
		if _, ok := results[ip]; !ok {
			results[ip] = false
		}
	}
	return results
}

func (s *ScannerService) StartScan(ipRange string) {
	ips, err := getIPList(ipRange)
	if err != nil {
		s.logger.Error("Invalid CIDR: ", ipRange)
		return
	}

	if s.cancel != nil {
		s.cancel()
		s.wg.Wait()
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.wg.Add(1)

	startedAt := time.Now()

	go func() {
		defer s.wg.Done()

		s.logger.Info("Scan started for range: ", ipRange)

		reachability := concurrentPing(ips, 1*time.Second)
		onlineCount := 0

		for _, ip := range ips {
			select {
			case <-ctx.Done():
				s.logger.Warn("Scan cancelled")
				return
			default:
				reachable := reachability[ip]
				if reachable {
					onlineCount++
				}
				hostname := ""
				mac := ""
				existing := s.repo.FindByIP(ip)
				var lastSeen time.Time

				if reachable {
					hostname = resolveHostname(ip)
					mac = resolveMAC(ip)
					lastSeen = time.Now()
					s.logger.Debug("Reachable: ", ip, " MAC: ", mac, " Hostname: ", hostname)
				} else if existing != nil {
					lastSeen = existing.LastSeen
					s.logger.Debug("Unreachable but exists: ", ip)
				}

				s.repo.Save(model.Device{
					IPAddress:  ip,
					MACAddress: mac,
					Hostname:   hostname,
					IsOnline:   reachable,
					LastSeen:   lastSeen,
				})
			}
		}

		if s.historyRepo != nil {
			if _, err := s.historyRepo.Save(model.ScanHistory{
				IPRange:     ipRange,
				StartedAt:   startedAt,
				DeviceCount: onlineCount,
			}); err != nil {
				s.logger.Error("Failed to persist scan history:", err)
			}
		}

		s.logger.Info("Scan completed for range: ", ipRange)
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
	s.logger.Info("All device records cleared.")
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
