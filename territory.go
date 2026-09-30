// Package territory 是领地占领：按在线人数与真实时间推进进度，无人争夺时衰减，严格超过阈值后翻转归属。
package territory

// FlipThreshold 是翻转阈值：进度严格超过 +FlipThreshold 归 a，严格低于 -FlipThreshold 归 b。
const FlipThreshold = 100.0

// RatePerSec 是每个净人数每秒推进的进度。
const RatePerSec = 10.0

// DecayPerSec 是无人争夺时每秒向 0 回落的进度。
const DecayPerSec = 2.0

// progressLimit 是进度的上下限。
const progressLimit = 100.0

// Point 是一个争夺点，Progress 大于 0 偏向 a，小于 0 偏向 b。
type Point struct {
	ID       string
	Owner    string
	Progress float64
}

// World 是领地世界。点保存在切片里、用 map 做 ID 到下标的索引：
// Tick/Owner/Progress 都是 O(1)，Snapshot 直接复制切片，从不扫描索引表。
type World struct {
	points  []Point
	index   map[string]int
	flips   int
	scanned int
}

// NewWorld 建一个世界，点按传入顺序保存。
func NewWorld(points []Point) *World {
	world := &World{
		points: make([]Point, len(points)),
		index:  make(map[string]int, len(points)),
	}
	for offset, point := range points {
		world.points[offset] = point
		world.index[point.ID] = offset
	}
	return world
}

// Tick 推进一个点：增量 = 净人数 * RatePerSec * dtMs/1000；双方都为 0 时按
// DecayPerSec * dtMs/1000 向 0 衰减。严格越过正负 FlipThreshold 时归属翻转、
// 进度清零、Flips 加一；进度始终夹在 [-100, 100]。
func (w *World) Tick(pointID string, aCount, bCount int, dtMs float64) {
	offset, ok := w.index[pointID]
	if !ok {
		return
	}
	point := &w.points[offset]
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

	switch {
	case point.Progress > FlipThreshold:
		point.Progress = 0
		point.Owner = "a"
		w.flips++
	case point.Progress < -FlipThreshold:
		point.Progress = 0
		point.Owner = "b"
		w.flips++
	case point.Progress > progressLimit:
		point.Progress = progressLimit
	case point.Progress < -progressLimit:
		point.Progress = -progressLimit
	}
}

// Owner 是当前归属。
func (w *World) Owner(pointID string) string {
	offset, ok := w.index[pointID]
	if !ok {
		return ""
	}
	return w.points[offset].Owner
}

// Progress 是当前进度。
func (w *World) Progress(pointID string) float64 {
	offset, ok := w.index[pointID]
	if !ok {
		return 0
	}
	return w.points[offset].Progress
}

// Flips 是累计翻转次数。
func (w *World) Flips() int { return w.flips }

// Snapshot 返回所有点（按构造顺序）的副本，只复制切片、不扫描索引表。
func (w *World) Snapshot() []Point {
	out := make([]Point, len(w.points))
	copy(out, w.points)
	return out
}

// Scanned 是累计扫描索引表的点数；快照走切片复制，因此不随点数或查询次数增长。
func (w *World) Scanned() int { return w.scanned }
