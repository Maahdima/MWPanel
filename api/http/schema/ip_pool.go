package schema

type IPPoolResponse struct {
	Id          uint   `json:"id"`
	Name        string `json:"name"`
	StartIP     string `json:"start_ip"`
	EndIP       string `json:"end_ip"`
	TotalIP     int    `json:"total_ip"`
	UsedIP      int    `json:"used_ip"`
	RemainingIP int    `json:"remaining_ip"`
}

type CreateIPPoolRequest struct {
	Name        string `json:"name" validate:"required"`
	InterfaceID uint   `json:"interface_id" validate:"required"`
	StartIP     string `json:"start_ip" validate:"required"`
	EndIP       string `json:"end_ip" validate:"required"`
}

type UpdateIPPoolRequest struct {
	Name    string `json:"name,omitempty"`
	StartIP string `json:"start_ip,omitempty"`
	EndIP   string `json:"end_ip,omitempty"`
}
