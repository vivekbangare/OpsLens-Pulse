package store

import "math"

func DetectZScore(points []TimelinePoint) bool {

	if len(points) < 10 {
		return false
	}

	window := 10
	recent := points[len(points)-window:]

	var sum float64
	for _, p := range recent {
		sum += float64(p.Total)
	}

	mean := sum / float64(len(recent))

	var variance float64
	for _, p := range recent {
		diff := float64(p.Total) - mean
		variance += diff * diff
	}

	stdDev := math.Sqrt(variance / float64(len(recent)))
	if stdDev == 0 {
		return false
	}

	last := float64(points[len(points)-1].Total)
	zScore := (last - mean) / stdDev

	return zScore > 3
}
