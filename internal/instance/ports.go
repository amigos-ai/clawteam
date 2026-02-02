package instance

import "fmt"

const (
	BasePort = 18789
	MaxPort  = 19000
)

func AllocatePort(reg *Registry) (int, error) {
	instances, err := reg.List()
	if err != nil {
		return 0, err
	}

	usedPorts := make(map[int]bool)
	for _, inst := range instances {
		usedPorts[inst.Port] = true
	}

	for port := BasePort; port < MaxPort; port++ {
		if !usedPorts[port] {
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available ports in range %d-%d", BasePort, MaxPort)
}
