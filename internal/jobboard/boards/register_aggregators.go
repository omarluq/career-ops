package boards

import (
	"github.com/omarluq/career-ops/internal/jobboard"
	"github.com/samber/lo"
)

// aggregatorBoards returns all aggregator and general job board instances.
func aggregatorBoards() []jobboard.Board {
	return []jobboard.Board{
		NewLinkedInBoard(),
		NewIndeedBoard(),
		NewGlassdoorBoard(),
		NewZipRecruiterBoard(),
		NewMonsterBoard(),
		NewCareerBuilderBoard(),
		NewDiceBoard(),
		NewUSAJobsBoard(),
		NewSnagajobBoard(),
		NewJoobleBoard(),
		NewAdzunaBoard(),
		NewTheMuseBoard(),
		NewJoblistBoard(),
		NewLaddersBoard(),
		NewRobertHalfBoard(),
		NewRandstadBoard(),
		NewReedBoard(),
		NewSeekBoard(),
		NewStepStoneBoard(),
	}
}

// RegisterAggregators adds all aggregator and general job boards to the registry.
func RegisterAggregators(reg *jobboard.Registry) {
	lo.ForEach(aggregatorBoards(), func(b jobboard.Board, _ int) {
		reg.Register(b)
	})
}
