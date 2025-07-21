package common

var (
	IPv4DefaultInterface = "ether1"
)

var (
	SchedulerComment   = "Expire WireGuard Peer: "
	SchedulerName      = "Schedule: "
	SchedulerStartTime = "12:00:00"
	SchedulerInterval  = "00:00:00"
	SchedulerPolicy    = "read,write"
	SchedulerEvent     = "/interface/wireguard/peers/disable"
)

var (
	QueueComment     = "Wg Bandwidth Queue: "
	QueueName        = "Bandwidth Limit: "
	DefaultKeepalive = "25"
)
