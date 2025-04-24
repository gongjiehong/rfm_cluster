// Package silhouette implements the silhouette cluster analysis algorithm
// See: https://en.wikipedia.org/wiki/Silhouette_(clustering)
package silhouette

import (
	"math"
	"rfm_cluster/pkg/clusters"
	"sync"
)

// KScore holds the score for a value of K
type KScore struct {
	Clusters clusters.Clusters
	K        int
	Score    float64
}

// Partitioner interface which suitable clustering algorithms should implement
type Partitioner interface {
	Partition(data clusters.Observations, k int) (clusters.Clusters, error)
}

// EstimateK estimates the amount of clusters (k) along with the silhouette
// score for that value, using the given partitioning algorithm
func EstimateK(data clusters.Observations, kmax int, m Partitioner) ([]KScore, int, float64, error) {
	scores, err := Scores(data, kmax, m)
	if err != nil {
		return nil, 0, -1.0, err
	}

	r := KScore{
		K: -1,
	}
	for _, score := range scores {
		if r.K < 0 || score.Score > r.Score {
			r = score
		}
	}

	return scores, r.K, r.Score, nil
}

// Scores calculates the silhouette scores for all values of k between 2 and
// kmax, using the given partitioning algorithm
func Scores(data clusters.Observations, kmax int, m Partitioner) ([]KScore, error) {
	var r []KScore = make([]KScore, kmax-1)

	waitGroup := sync.WaitGroup{}
	for k := 2; k <= kmax; k++ {
		index := k - 2
		waitGroup.Add(1)
		go func(index int) {
			defer waitGroup.Done()
			cc, s, err := Score(data, k, m)
			if err != nil {
				panic(err)
			}

			r[index] = KScore{
				Clusters: cc,
				K:        k,
				Score:    s,
			}
		}(index)
	}

	waitGroup.Wait()

	return r, nil
}

// Score calculates the silhouette score for a given value of k, using the given
// partitioning algorithm
func Score(data clusters.Observations, k int, m Partitioner) (clusters.Clusters, float64, error) {
	cc, err := m.Partition(data, k)
	if err != nil {
		return cc, -1.0, err
	}

	var si float64
	var sc int64
	for ci, c := range cc {
		for _, p := range c.Observations {
			ai := clusters.AverageDistance(p, c.Observations)
			_, bi := cc.Neighbour(p, ci)

			si += (bi - ai) / math.Max(ai, bi)
			sc++
		}
	}

	return cc, si / float64(sc), nil
}
