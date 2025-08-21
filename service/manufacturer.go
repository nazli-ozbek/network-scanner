package service

type ManufacturerResolver interface {
	Resolve(mac string) string
}

type MockManufacturerResolver struct{}

func (m *MockManufacturerResolver) Resolve(mac string) string {
	if len(mac) < 8 {
		return ""
	}
	p := mac[:8]
	switch p {
	case "AA:BB:CC":
		return "A"
	default:
		return ""
	}
}
