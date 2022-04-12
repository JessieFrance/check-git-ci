package checkgitci

// NewRateManager returns a RateManager struct.
func NewRateManager() RateManager {
	return RateManager{
		Remaining:   make(map[string]int),
		APIKeys:     make(map[string]string),
		CallHeaders: make(map[string]APIHeaders),
	}
}
