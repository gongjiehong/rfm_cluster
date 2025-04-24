// Package kmeans implements the k-means clustering algorithm
// See: https://en.wikipedia.org/wiki/K-means_clustering
package kmeans

import (
	"fmt"
	"math"
	"math/rand"
	"rfm_cluster/pkg/clusters"
	"time"
)

// Kmeans configuration/option struct
type Kmeans struct {
	// when a plotter is set, Plot gets called after each iteration
	plotter Plotter
	// deltaThreshold (in percent between 0.0 and 0.1) aborts processing if
	// less than n% of data points shifted clusters in the last iteration
	deltaThreshold float64
	// iterationThreshold aborts processing when the specified amount of
	// algorithm iterations was reached
	iterationThreshold int
}

// The Plotter interface lets you implement your own plotters
type Plotter interface {
	Plot(cc clusters.Clusters, iteration int) error
}

// NewWithOptions returns a Kmeans configuration struct with custom settings
func NewWithOptions(deltaThreshold float64, plotter Plotter) (Kmeans, error) {
	if deltaThreshold <= 0.0 || deltaThreshold >= 1.0 {
		return Kmeans{}, fmt.Errorf("threshold is out of bounds (must be >0.0 and <1.0, in percent)")
	}

	return Kmeans{
		plotter:            plotter,
		deltaThreshold:     deltaThreshold,
		iterationThreshold: 96,
	}, nil
}

// New returns a Kmeans configuration struct with default settings
func New() Kmeans {
	m, _ := NewWithOptions(0.01, nil)
	return m
}

// initializeClustersKmeansPP 使用k-means++算法初始化聚类中心
func initializeClustersKmeansPP(k int, dataset clusters.Observations) (clusters.Clusters, error) {
	if k > len(dataset) {
		return clusters.Clusters{}, fmt.Errorf("the size of the data set must at least equal k")
	}

	// 创建k个空集群
	cc := make(clusters.Clusters, k)

	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	// 随机选择第一个聚类中心
	// firstCenterIdx := r.Intn(len(dataset))
	// 固定一个中心
	firstCenterIdx := len(dataset) / 2
	cc[0].Center = dataset[firstCenterIdx].Coordinates()

	// 选择剩余的k-1个聚类中心
	for i := 1; i < k; i++ {
		// 计算每个点到最近聚类中心的距离的平方
		distSquared := make([]float64, len(dataset))
		sumDistSquared := 0.0

		for j, point := range dataset {
			// 找到最近的聚类中心
			minDist := math.MaxFloat64
			for c := 0; c < i; c++ {
				dist := point.Distance(cc[c].Center)
				if dist < minDist {
					minDist = dist
				}
			}
			distSquared[j] = minDist
			sumDistSquared += minDist
		}

		// 使用距离的平方作为权重选择下一个中心点
		targetValue := r.Float64() * sumDistSquared
		currentSum := 0.0
		nextCenterIdx := 0

		// 根据权重概率选择下一个中心点
		for j, dist := range distSquared {
			currentSum += dist
			if currentSum >= targetValue {
				nextCenterIdx = j
				break
			}
		}

		cc[i].Center = dataset[nextCenterIdx].Coordinates()
	}

	return cc, nil
}

// Partition executes the k-means algorithm on the given dataset and
// partitions it into k clusters
func (m Kmeans) Partition(dataset clusters.Observations, k int) (clusters.Clusters, error) {
	if k > len(dataset) {
		return clusters.Clusters{}, fmt.Errorf("the size of the data set must at least equal k")
	}

	// 使用k-means++算法初始化聚类中心
	cc, err := initializeClustersKmeansPP(k, dataset)
	if err != nil {
		return clusters.Clusters{}, err
	}

	points := make([]int, len(dataset))
	changes := 1

	for i := 0; changes > 0; i++ {
		changes = 0
		cc.Reset()

		for p, point := range dataset {
			ci := cc.Nearest(point)
			cc[ci].Append(point)
			if points[p] != ci {
				points[p] = ci
				changes++
			}
		}

		for ci := 0; ci < len(cc); ci++ {
			if len(cc[ci].Observations) == 0 {
				// During the iterations, if any of the cluster centers has no
				// data points associated with it, assign a random data point
				// to it.
				// Also see: http://user.ceng.metu.edu.tr/~tcan/ceng465_f1314/Schedule/KMeansEmpty.html
				var ri int
				for {
					// find a cluster with at least two data points, otherwise
					// we're just emptying one cluster to fill another
					ri = rand.Intn(len(dataset)) //nolint:gosec // rand.Intn is good enough for this
					if len(cc[points[ri]].Observations) > 1 {
						break
					}
				}
				cc[ci].Append(dataset[ri])
				points[ri] = ci

				// Ensure that we always see at least one more iteration after
				// randomly assigning a data point to a cluster
				changes = len(dataset)
			}
		}

		if changes > 0 {
			cc.Recenter()
		}
		if m.plotter != nil {
			err := m.plotter.Plot(cc, i)
			if err != nil {
				return nil, fmt.Errorf("failed to plot chart: %s", err)
			}
		}
		if i == m.iterationThreshold ||
			changes < int(float64(len(dataset))*m.deltaThreshold) {
			// fmt.Println("Aborting:", changes, int(float64(len(dataset))*m.TerminationThreshold))
			break
		}
	}

	return cc, nil
}
