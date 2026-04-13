package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// RegisterAll registers all 100 boards with the given registry.
func RegisterAll(reg *jobboard.Registry) {
	RegisterATS(reg)
	RegisterAggregators(reg)
	RegisterStartup(reg)
	RegisterRemote(reg)
	RegisterFreelance(reg)
}
