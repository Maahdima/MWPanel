package schema

type PeerSessionStatus string

var (
	PeerSessionOngoing PeerSessionStatus = "ongoing"
	PeerSessionEnded   PeerSessionStatus = "ended"
)

type PeerSessionResponse struct {
	ID             uint              `json:"id"`
	Status         PeerSessionStatus `json:"status"`
	ConnectedAt    string            `json:"connected_at"`
	DisconnectedAt *string           `json:"disconnected_at"`
	Duration       int64             `json:"duration"`
	DurationLabel  string            `json:"duration_label"`
	Endpoint       *string           `json:"endpoint"`
	DownloadUsage  string            `json:"download_usage"`
	UploadUsage    string            `json:"upload_usage"`
	TotalUsage     string            `json:"total_usage"`
}

type PeerSessionsResponse struct {
	PeerID             uint                  `json:"peer_id"`
	PeerName           string                `json:"peer_name"`
	TotalSessions      int                   `json:"total_sessions"`
	OngoingSessions    int                   `json:"ongoing_sessions"`
	TotalDuration      int64                 `json:"total_duration"`
	TotalDurationLabel string                `json:"total_duration_label"`
	TotalDownloadUsage string                `json:"total_download_usage"`
	TotalUploadUsage   string                `json:"total_upload_usage"`
	TotalUsage         string                `json:"total_usage"`
	Sessions           []PeerSessionResponse `json:"sessions"`
}
