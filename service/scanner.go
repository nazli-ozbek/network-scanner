package service

import (
	"context"
	"net"
	"network-scanner/logger"
	"network-scanner/model"
	"network-scanner/repository"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/j-keck/arping"
	"github.com/tatsushid/go-fastping"
)

type ScannerService struct {
	repo       repository.DeviceRepository
	cancel     context.CancelFunc
	logger     logger.Logger
	wg         sync.WaitGroup
	resolver   ManufacturerResolver
	pollCancel context.CancelFunc
	pollWG     sync.WaitGroup
}

func NewScannerService(repo repository.DeviceRepository, logger logger.Logger) *ScannerService {
	return &ScannerService{repo: repo, logger: logger, resolver: nil}
}

func NewScannerServiceWithResolver(repo repository.DeviceRepository, logger logger.Logger, r ManufacturerResolver) *ScannerService {
	return &ScannerService{repo: repo, logger: logger, resolver: r}
}

func concurrentPing(ips []string, timeout time.Duration) map[string]bool {
	p := fastping.NewPinger()
	p.MaxRTT = timeout
	results := make(map[string]bool)
	var mu sync.Mutex
	for _, ip := range ips {
		if addr, err := net.ResolveIPAddr("ip4:icmp", ip); err == nil {
			p.AddIPAddr(addr)
		}
	}
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		mu.Lock()
		results[addr.String()] = true
		mu.Unlock()
	}
	p.OnIdle = func() {}
	_ = p.Run()
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

	go func() {
		defer s.wg.Done()
		s.logger.Info("Scan started for range: ", ipRange)
		reachability := concurrentPing(ips, 1*time.Second)

		for _, ip := range ips {
			select {
			case <-ctx.Done():
				s.logger.Warn("Scan cancelled")
				return
			default:
				reachable := reachability[ip]
				existing := s.repo.FindByIP(ip)

				var (
					id           string
					hostname     string
					mac          string
					lastSeen     time.Time
					firstSeen    time.Time
					manufacturer string
					tags         []string
				)

				if existing != nil {
					id = existing.ID
					firstSeen = existing.FirstSeen
					manufacturer = existing.Manufacturer
					tags = existing.Tags
				} else {
					id = uuid.New().String()
				}

				status := "offline"
				if reachable {
					status = "online"
					hostname = resolveHostname(ip)
					mac = resolveMAC(ip)
					lastSeen = time.Now()

					if firstSeen.IsZero() {
						firstSeen = lastSeen
					}

					if s.resolver != nil && mac != "" {
						if resolved := s.resolver.Resolve(mac); resolved != "" {
							manufacturer = resolved
						}
					}
				} else if existing != nil {
					lastSeen = existing.LastSeen
					hostname = existing.Hostname
					mac = existing.MACAddress
				}

				device := model.Device{
					ID:           id,
					IPAddress:    ip,
					MACAddress:   mac,
					Hostname:     hostname,
					Status:       status,
					Manufacturer: manufacturer,
					Tags:         tags,
					LastSeen:     lastSeen,
					FirstSeen:    firstSeen,
				}

				s.repo.Save(device)
			}
		}
		s.logger.Info("Scan completed for range: ", ipRange)
	}()
}

func (s *ScannerService) StartStatusPolling(interval time.Duration) {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	if s.pollCancel != nil {
		s.pollCancel()
		s.pollWG.Wait()
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.pollCancel = cancel
	s.pollWG.Add(1)
	go func() {
		defer s.pollWG.Done()
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				devs := s.repo.GetAll()
				if len(devs) == 0 {
					continue
				}
				ips := make([]string, 0, len(devs))
				for _, d := range devs {
					if d.IPAddress != "" {
						ips = append(ips, d.IPAddress)
					}
				}
				reach := concurrentPing(ips, 1*time.Second)
				now := time.Now()
				for _, d := range devs {
					if reach[d.IPAddress] {
						if d.FirstSeen.IsZero() {
							d.FirstSeen = now
						}
						d.LastSeen = now
						d.Status = "online"
					} else {
						d.Status = "offline"
					}
					s.repo.Save(d)
				}
			}
		}
	}()
}

func (s *ScannerService) UpdateTags(id string, tags []string) error {
	norm := normalizeTags(tags, 10)
	return s.repo.UpdateTags(id, norm)
}

func (s *ScannerService) FindByID(id string) (*model.Device, error) {
	return s.repo.FindByID(id)
}

func (s *ScannerService) SearchDevices(q string) ([]model.Device, error) {
	return s.repo.Search(q)
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
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) { reachable = true }
	if err = p.Run(); err != nil {
		return false
	}
	return reachable
}

func getIPList(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	var out []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		out = append(out, ip.String())
	}
	if len(out) > 2 {
		out = out[1 : len(out)-1]
	}
	return out, nil
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
