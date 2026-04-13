package boards

import (
	"github.com/samber/lo"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// RegisterFreelance registers all freelance and international boards with the given registry.
func RegisterFreelance(reg *jobboard.Registry) {
	lo.ForEach(freelanceBoards(), func(b jobboard.Board, _ int) { reg.Register(b) })
}
