package monitor

// SuppressedCount returns the total number of alerts that have been suppressed
// (i.e., calls to ShouldSend that returned false due to cooldown).
//
// This method extends DedupStore to support summary reporting without
// breaking the existing DedupStore API.
func (d *DedupStore) SuppressedCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.suppressedCount
}
