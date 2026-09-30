// Package territory 是领地占领：按在线人数与真实时间推进进度，无人争夺时衰减，过阈值翻转归属。
// 缺陷：进度按调用次数推进、无人争夺不衰减、到阈值不翻转、进度不夹范围、查询全表扫描。
package territory

// FlipThreshold 是翻转阈值。
const FlipThreshold = 100.0

// RatePerSec 是每个净人数每秒推进的进度。
const RatePerSec = 10.0

// DecayPerSec 是无人争夺时每秒回落的进度。
const DecayPerSec = 2.0

// Point 是一个争夺点，Progress 大于 0 偏向 a，小于 0 偏向 b。
type Point struct {
	ID       string
	Owner    string
	Progress float64
}

// World 是领地世界。
type World struct {
	points  map[string]*Point
	flips   int
	scanned int
}

// NewWorld 建一个世界。
func NewWorld(points []Point) *World {
	world := &World{points: map[string]*Point{}}
	for index := range points {
		point := points[index]
		world.points[point.ID] = &point
	}
	return world
}

// Tick 推进一个点。
// 缺陷：进度按调用次数加（不看 dtMs）、无人争夺不衰减、到阈值不翻转、进度不夹范围。
func (w *World) Tick(pointID string, aCount, bCount int, dtMs float64) {
	point := w.points[pointID]
	if point == nil {
		return
	}
	point.Progress += float64(aCount - bCount)
}

// Owner 是当前归属。
func (w *World) Owner(pointID string) string {
	point := w.points[pointID]
	if point == nil {
		return ""
	}
	return point.Owner
}

// Progress 是当前进度。
func (w *World) Progress(pointID string) float64 {
	point := w.points[pointID]
	if point == nil {
		return 0
	}
	return point.Progress
}

// Flips 是累计翻转次数。
func (w *World) Flips() int { return w.flips }

// Snapshot 返回所有点。缺陷：每次全表扫描。
func (w *World) Snapshot() []Point {
	out := make([]Point, 0, len(w.points))
	for _, point := range w.points {
		w.scanned++
		out = append(out, *point)
	}
	return out
}

// Scanned 是累计扫描点数（规模观测）。
func (w *World) Scanned() int { return w.scanned }
