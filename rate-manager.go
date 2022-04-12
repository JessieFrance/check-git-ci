package checkgitci

// NewRateManager returns a RateManager struct.
func NewRateManager() RateManager {
	return RateManager{
		Remaining:   make(map[string]int),
		APIKeys:     make(map[string]string),
		CallHeaders: make(map[string]APIHeaders),
	}
}

// SetKey sets a general key intended for a single RateManager,
// that may be used for multiple repositories.
func (r *RateManager) SetKey(key string) {
	r.APIKeys[generalKey] = key
}
