package utils

// PeerCounterDelta shares the existing RouterOS wrap/reset policy between jobs and exact billing.
func PeerCounterDelta(prev, current, maxCounter int64) (int64, bool) {
	if current >= prev {
		return current - prev, false
	}

	// If counters are already beyond the expected 32-bit max, treat this as a reset.
	if maxCounter <= 0 || prev > maxCounter || current > maxCounter {
		return current, true
	}

	// If the counter moved backwards by a large amount, assume a wrap.
	if (prev - current) > (maxCounter / 2) {
		return (maxCounter - prev) + current, false
	}

	// Otherwise treat it as a reset to avoid a large, incorrect delta.
	return current, true
}
