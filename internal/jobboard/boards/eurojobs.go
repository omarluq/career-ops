package boards

import "github.com/omarluq/career-ops/internal/jobboard"

// NewEuroJobsBoard creates a EuroJobs board (ScrapeBoard).
func NewEuroJobsBoard() *ScrapeBoard {
	return NewScrapeBoard(jobboard.BoardMeta{
		URL:          "https://www.eurojobs.com",
		Name:         "EuroJobs",
		Slug:         "eurojobs",
		Category:     jobboard.CategoryAggregator,
		AuthType:     jobboard.AuthNone,
		RateLimit:    jobboard.RateConfig{RequestsPerMinute: 20, BurstSize: 3},
		Capabilities: []jobboard.Capability{jobboard.CapSearch, jobboard.CapScrape},
	}, "https://www.eurojobs.com/search", "/job/")
}

// ---------------------------------------------------------------------------
// freelanceBoards returns all boards in the freelance + international batch.
// ---------------------------------------------------------------------------

// freelanceBoards returns all 20 freelance and international board instances.
func freelanceBoards() []jobboard.Board {
	return []jobboard.Board{
		NewToptalBoard(),
		NewUpworkBoard(),
		NewFreelancerBoard(),
		NewFiverrBoard(),
		NewGunIOBoard(),
		NewContraBoard(),
		NewArcBoard(),
		NewTuringBoard(),
		NewAndelaBoard(),
		NewCrossoverBoard(),
		NewNaukriBoard(),
		NewJobsDBBoard(),
		NewBaytBoard(),
		NewXingBoard(),
		NewTotaljobsBoard(),
		NewHaysBoard(),
		NewJobbermanBoard(),
		NewWantedlyBoard(),
		NewCakeResumeBoard(),
		NewEuroJobsBoard(),
	}
}
