package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// startupBoards returns all 20 startup and niche tech board instances.
func startupBoards() []jobboard.Board {
	return []jobboard.Board{
		NewYCombinatorBoard(),
		NewWellfoundBoard(),
		NewUnderdogBoard(),
		NewBuiltInBoard(),
		NewVentureLoopBoard(),
		NewWorkAtAStartup2Board(),
		NewFundedJobsBoard(),
		NewTechStarsBoard(),
		NewProductHuntBoard(),
		NewF6SBoard(),
		NewStartupJobsBoard(),
		NewLemonBoard(),
		NewPalletBoard(),
		NewTrueUpBoard(),
		NewUntappedBoard(),
	}
}
