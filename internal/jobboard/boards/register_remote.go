package boards

import (
	"github.com/samber/lo"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// RegisterRemote registers all remote, AI, and specialty boards with the given registry.
func RegisterRemote(reg *jobboard.Registry) {
	lo.ForEach(remoteBoards(), func(b jobboard.Board, _ int) { reg.Register(b) })
}
