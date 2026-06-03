package generators

import "github.com/gouef/githubtoplanguages/requests"

type StreaksSection struct {
	Data   *requests.StreakStats
	Height int
}

func NewStreaksSection(streak *requests.StreakStats, show bool) *StreaksSection {
	if !show || streak == nil {
		return &StreaksSection{Height: 0}
	}

	return &StreaksSection{
		Data:   streak,
		Height: 130,
	}
}
