package entities

import (
	"fmt"
	"proxy-service/application/config"
	"sync"
)

var (
	proxy     *Proxy
	proxyOnce sync.Once
)

// GetProxy retrieves the Proxy configuration.
func GetProxy() *Proxy {
	proxyOnce.Do(func() {
		cfg := config.GetConfig()
		proxy = &Proxy{
			Host: cfg.Proxy.Host,
			Port: cfg.Proxy.Port,
		}
	})
	return proxy
}

// Proxy represents the details of a proxy configuration.
type Proxy struct {
	Host string // Host is the hostname of the proxy server.
	Port string // Port is the port number of the proxy server.
}

// Address returns the full address of the proxy.
func (p *Proxy) Address() (uri string, err error) {
	if p.Host == "" || p.Port == "" {
		return "", fmt.Errorf("invalid address: host or port is empty")
	}
	return fmt.Sprintf("%s:%s", p.Host, p.Port), nil
}
