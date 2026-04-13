package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// remoteBoards returns all remote, AI, and specialty board instances.
func remoteBoards() []jobboard.Board {
	return []jobboard.Board{
		NewRemoteOKBoard(),
		NewWeWorkRemotelyBoard(),
		NewFlexJobsBoard(),
		NewRemoteCoBoard(),
		NewRemotiveBoard(),
		NewJustRemoteBoard(),
		NewDailyRemoteBoard(),
		NewRemoteLeafBoard(),
		NewAIJobsNetBoard(),
		NewAIJobBoardBoard(),
		NewMLJobListBoard(),
		NewAIJobsIOBoard(),
		NewDataScienceJobsBoard(),
		NewHackerNewsJobsBoard(),
		NewCryptoJobsListBoard(),
		NewWeb3CareerBoard(),
		NewClimateBaseBoard(),
	}
}
