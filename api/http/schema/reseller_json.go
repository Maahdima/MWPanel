package schema

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// Optional int64 settings accept integer strings so browser clients never need
// floating-point conversion for prices or byte quantities.
func optionalInteger(data json.RawMessage) (*int64, error) {
	if len(data) == 0 || bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return nil, nil
	}

	text := string(data)

	if len(data) > 0 && data[0] == '"' {
		if err := json.Unmarshal(data, &text); err != nil {
			return nil, err
		}
	}

	value, err := strconv.ParseInt(text, 10, 64)

	if err != nil {
		return nil, err
	}

	return &value, nil
}

func (s *ResellerSettings) UnmarshalJSON(data []byte) error {
	type plain ResellerSettings

	var fields map[string]json.RawMessage

	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	keys := []string{"rate_rial", "max_active_peers", "max_peer_traffic_bytes", "max_total_traffic_bytes", "max_duration_days", "interval_days", "grace_days", "recurring_fee_rial"}

	saved := make([]json.RawMessage, len(keys))

	for index, key := range keys {
		saved[index] = fields[key]

		delete(fields, key)
	}

	remaining, err := json.Marshal(fields)

	if err != nil {
		return err
	}

	var result plain

	if err := json.Unmarshal(remaining, &result); err != nil {
		return err
	}

	targets := []**int64{&result.RateRial, &result.MaxActivePeers, &result.MaxPeerTrafficBytes, &result.MaxTotalTrafficBytes, &result.MaxDurationDays, &result.IntervalDays, &result.GraceDays, &result.RecurringFeeRial}

	for index, target := range targets {
		value, err := optionalInteger(saved[index])

		if err != nil {
			return err
		}

		*target = value
	}

	*s = ResellerSettings(result)

	return nil
}

func (s *CreateReseller) UnmarshalJSON(data []byte) error {
	var login struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.Unmarshal(data, &login); err != nil {
		return err
	}

	if err := json.Unmarshal(data, &s.ResellerSettings); err != nil {
		return err
	}

	s.Username = login.Username

	s.Password = login.Password

	return nil
}
