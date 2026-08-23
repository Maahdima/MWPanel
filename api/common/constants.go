package common

import "time"

const PeerOnlineHandshakeTimeout = 150 * time.Second

var (
	DeviceInfoPath     = "/system/resource"
	DeviceIdentityPath = "/system/identity"
	DeviceDnsPath      = "/ip/dns"
	DeviceIPv4Path     = "/ip/address"
	InterfacePath      = "/interface"
	WGInterfacePath    = "/interface/wireguard"
	WGPeerPath         = "/interface/wireguard/peers"
	QueuePath          = "/queue/simple"
	SchedulerPath      = "/system/scheduler"
)

var (
	IPv4DefaultInterface = "ether1"
)

var (
	SchedulerComment   = "Expire WireGuard Peer: "
	SchedulerName      = "Schedule: "
	SchedulerStartTime = "00:00:00"
	SchedulerInterval  = "00:00:00"
	SchedulerPolicy    = "read,write"
	SchedulerEvent     = `/interface/wireguard/peers/set [find where .id="%s"] disabled=yes`
)

var (
	QueueComment     = "Wireguard Bandwidth Queue: "
	QueueName        = "Bandwidth Limit: "
	DefaultKeepalive = "25"
)

// TODO : Make these configurable
var (
	AllowedIpsExcludeLocal = ""
	AllowedIpsIncludeLocal = "0.0.0.0/0, ::/0"
	DefaultDns             = "8.8.8.8, 1.1.1.1"
)
