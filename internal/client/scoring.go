package client

import (
	"cmp"
	"slices"

	"github.com/sverrehu/spacegame/internal/model"
)

type Score struct {
	Name      string
	Score     int16
	AntiScore int16
	Ratio     float64
}

func toScores(ships []*model.Ship) []Score {
	scores := make([]Score, len(ships))
	for i, ship := range ships {
		scores[i] = Score{Name: ship.Name, Score: ship.Score, AntiScore: ship.AntiScore}
		scores[i].Ratio = calcRatio(ship.Score, ship.AntiScore)
	}
	slices.SortFunc(scores, func(s0 Score, s1 Score) int {
		if s0.Ratio == s1.Ratio {
			return cmp.Compare(s0.Name, s1.Name)
		}
		return cmp.Compare(s0.Ratio, s1.Ratio)
	})
	return scores
}

func calcRatio(score int16, antiScore int16) float64 {
	// algorithm by Jon S. Bratseth back in 1998
	if antiScore == 0 {
		return float64(score) * 2.0
	}
	return float64(score) / float64(antiScore)
}
