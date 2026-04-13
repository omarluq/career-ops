package scanner

import "strings"

// DefaultBoards returns all built-in board adapters.
func DefaultBoards() []BoardAdapter {
	return []BoardAdapter{
		&LinkedInAdapter{},
		&YCombinatorAdapter{},
		&WellfoundAdapter{},
		&AIJobsAdapter{},
	}
}

// BoardByName returns the subset of boards matching the given names (case-insensitive).
// If names is empty, all boards are returned.
func BoardByName(boards []BoardAdapter, names ...string) []BoardAdapter {
	if len(names) == 0 {
		return boards
	}

	nameSet := make(map[string]struct{}, len(names))
	for _, n := range names {
		nameSet[strings.ToLower(n)] = struct{}{}
	}

	var matched []BoardAdapter
	for _, b := range boards {
		if _, ok := nameSet[strings.ToLower(b.Name())]; ok {
			matched = append(matched, b)
		}
	}

	return matched
}
