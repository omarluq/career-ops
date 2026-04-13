package jobboard

import "errors"

var (
	// ErrApplyNotSupported is returned when a board does not support application submission.
	ErrApplyNotSupported = errors.New("board does not support apply")
	// ErrRateLimited is returned when the board's rate limit has been exceeded.
	ErrRateLimited = errors.New("rate limited by board")
	// ErrAuthRequired is returned when the board requires authentication credentials.
	ErrAuthRequired = errors.New("authentication required")
	// ErrBoardUnreachable is returned when a board cannot be contacted.
	ErrBoardUnreachable = errors.New("board is unreachable")
)
