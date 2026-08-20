package config

import (
	"fmt"
	"net"
)

// cgnatMesh is the 100.64.0.0/10 CGNAT range commonly used by overlay meshes
// (Nebula/Tailscale-style). It is treated as a trusted mesh range.
var cgnatMesh = &net.IPNet{IP: net.IPv4(100, 64, 0, 0), Mask: net.CIDRMask(10, 32)}

// CheckBindAddress returns an error if host is a public address that the server
// must not bind without an explicit override. Loopback, RFC1918/RFC4193
// private, CGNAT-mesh, and link-local addresses are allowed by default because
// the server is reached only via the loopback interface or the Nebula mesh
// (TLS is terminated upstream by the Hetzner proxy).
//
// If override is true, any address is permitted (the "--i-know-what-im-doing"
// escape hatch).
func CheckBindAddress(host string, override bool) error {
	if override {
		return nil
	}
	if host == "localhost" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("refusing to bind %q: not a verifiable IP; pass the override to bind a hostname", host)
	}
	switch {
	case ip.IsLoopback():
		return nil
	case ip.IsPrivate():
		return nil
	case ip.IsLinkLocalUnicast():
		return nil
	case cgnatMesh.Contains(ip):
		return nil
	case ip.IsUnspecified():
		// 0.0.0.0 / :: bind everything — that includes public interfaces.
		return fmt.Errorf("refusing to bind the unspecified address %q (exposes public interfaces); bind loopback/mesh or pass the override", host)
	default:
		return fmt.Errorf("refusing to bind public address %q; TLS terminates upstream — bind loopback/mesh or pass the override", host)
	}
}
