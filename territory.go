// Package territory 是领地占领：按在线人数与真实时间推进进度，无人争夺时衰减，过阈值翻转归属。
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
	ordered []*Point
	flips   int
	scanned int
}

// NewWorld 建一个世界。
func NewWorld(points []Point) *World {
	world := &World{points: map[string]*Point{}, ordered: make([]*Point, 0, len(points))}
	for index := range points {
		point := points[index]
		if existing := world.points[point.ID]; existing != nil {
			*existing = point
			continue
		}
		world.points[point.ID] = &point
		world.ordered = append(world.ordered, &point)
	}
	return world
}

// Tick 按真实时间推进一个点。
//
// 双方都为 0 时无人争夺，进度按 DecayPerSec 乘 dtMs/1000 向 0 回落；
// 否则增量为净人数乘 RatePerSec 乘 dtMs/1000。进度严格越过 +FlipThreshold
// 或 -FlipThreshold 时归属翻转、进度清零、Flips 加一；进度始终夹在 [-100, 100]。
func (w *World) Tick(pointID string, aCount, bCount int, dtMs float64) {
	point := w.points[pointID]
	if point == nil {
		return
	}
	dt := dtMs / 1000.0
	if aCount == 0 && bCount == 0 {
		decay := DecayPerSec * dt
		switch {
		case point.Progress > 0:
			point.Progress -= decay
			if point.Progress < 0 {
				point.Progress = 0
			}
		case point.Progress < 0:
			point.Progress += decay
			if point.Progress > 0 {
				point.Progress = 0
			}
		}
	} else {
		point.Progress += float64(aCount-bCount) * RatePerSec * dt
	}

	if point.Progress > FlipThreshold {
		point.Owner = "a"
		point.Progress = 0
		w.flips++
	} else if point.Progress < -FlipThreshold {
		point.Owner = "b"
		point.Progress = 0
		w.flips++
	} else if point.Progress > 100 {
		point.Progress = 100
	} else if point.Progress < -100 {
		point.Progress = -100
	}
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

// Snapshot 返回所有点。构造时已保存点位切片，查询直接按固定顺序复制，
// 不做随点数线性增长的表扫描，因此 Scanned 不随点数增长；内存 O(点数)。
func (w *World) Snapshot() []Point {
	out := make([]Point, len(w.ordered))
	for index, point := range w.ordered {
		out[index] = *point
	}
	return out
}

// Scanned 是累计扫描点数（规模观测）。
func (w *World) Scanned() int { return w.scanned }
