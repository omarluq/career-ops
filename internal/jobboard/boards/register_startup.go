package boards

import (
	"github.com/samber/lo"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// RegisterStartup registers all startup and niche tech boards with the given registry.
func RegisterStartup(reg *jobboard.Registry) {
	lo.ForEach(startupBoards(), func(b jobboard.Board, _ int) { reg.Register(b) })
}
