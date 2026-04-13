package boards

import (
	"github.com/samber/lo"

	"github.com/omarluq/career-ops/internal/jobboard"
)

// RegisterATS registers all ATS board implementations with the given registry.
func RegisterATS(reg *jobboard.Registry) {
	lo.ForEach(atsBoards(), func(b jobboard.Board, _ int) {
		reg.Register(b)
	})
}

// atsBoards returns all 20 ATS board implementations.
func atsBoards() []jobboard.Board {
	return []jobboard.Board{
		NewGreenhouseBoard(),
		NewLeverBoard(),
		NewAshbyBoard(),
		NewWorkableBoard(),
		NewBambooHRBoard(),
		NewICIMSBoard(),
		NewJobviteBoard(),
		NewSmartRecruitersBoard(),
		NewWorkdayBoard(),
		NewTaleoBoard(),
		NewJazzHRBoard(),
		NewRecruiteeBoard(),
		NewBreezyBoard(),
		NewPinpointBoard(),
		NewPersonioBoard(),
		NewTeamTailorBoard(),
		NewDoverBoard(),
		NewRipplingBoard(),
		NewFountainBoard(),
		NewApplicantProBoard(),
	}
}
