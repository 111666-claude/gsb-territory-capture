package territory

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

// tickByTime 是「按时间推进 + 衰减 + 阈值翻转」的独立参考实现，
// 用与生产代码不同的写法直接套口径，供差分测试逐点比对。
type refState struct {
	progress float64
	owner    string
	flips    int
}

func refTick(state *refState, aCount, bCount int, dtMs float64) {
	seconds := dtMs / 1000.0
	delta := float64(aCount-bCount) * RatePerSec * seconds
	if aCount == 0 && bCount == 0 {
		magnitude := math.Abs(state.progress) - DecayPerSec*seconds
		if magnitude < 0 {
			magnitude = 0
		}
		state.progress = math.Copysign(magnitude, state.progress)
	} else {
		state.progress += delta
	}
	if state.progress > FlipThreshold {
		state.owner = "a"
		state.progress = 0
		state.flips++
	} else if state.progress < -FlipThreshold {
		state.owner = "b"
		state.progress = 0
		state.flips++
	}
}

func TestAdvanceByRealTime(t *testing.T) {
	instance := world()
	instance.Tick("gate", 1, 0, 1000)
	if got := instance.Progress("gate"); got != 10 {
		t.Fatalf("一人 1 秒应推进 10，实际 %v", got)
	}

	long := world()
	long.Tick("gate", 1, 0, 10000)
	if got := long.Progress("gate"); got != 100 {
		t.Fatalf("一人 10 秒应推进 100，实际 %v", got)
	}

	tenCalls := world()
	for index := 0; index < 10; index++ {
		tenCalls.Tick("gate", 1, 0, 100)
	}
	if got := tenCalls.Progress("gate"); got != 10 {
		t.Fatalf("进度只与真实时间有关、与调用次数无关，实际 %v", got)
	}

	netZero := world()
	netZero.Tick("gate", 3, 3, 5000)
	if got := netZero.Progress("gate"); got != 0 {
		t.Fatalf("净人数为 0 不该推进，实际 %v", got)
	}
}

func TestDecayTowardZeroWithoutContest(t *testing.T) {
	forward := world()
	forward.Tick("gate", 4, 0, 1000)
	forward.Tick("gate", 0, 0, 10000)
	if got := forward.Progress("gate"); got != 20 {
		t.Fatalf("无人争夺应衰减到 20，实际 %v", got)
	}

	clamp := world()
	clamp.Tick("gate", 4, 0, 1000)
	clamp.Tick("gate", 0, 0, 100000)
	if got := clamp.Progress("gate"); got != 0 {
		t.Fatalf("衰减不能越过 0，实际 %v", got)
	}
	if clamp.Owner("gate") != "" {
		t.Fatal("单纯衰减不该改变归属")
	}

	backward := world()
	backward.Tick("gate", 0, 4, 1000)
	backward.Tick("gate", 0, 0, 1000)
	if got := backward.Progress("gate"); got != -38 {
		t.Fatalf("负进度应向 0 衰减，实际 %v", got)
	}
}

func TestFlipAtThreshold(t *testing.T) {
	aFlip := world()
	aFlip.Tick("gate", 12, 0, 1000)
	if got := aFlip.Progress("gate"); got != 0 {
		t.Fatalf("翻转后进度应清零，实际 %v", got)
	}
	if aFlip.Owner("gate") != "a" || aFlip.Flips() != 1 {
		t.Fatalf("越过 +阈值应翻给 a 并计一次，owner=%q flips=%d", aFlip.Owner("gate"), aFlip.Flips())
	}

	bFlip := world()
	bFlip.Tick("gate", 0, 12, 1000)
	if bFlip.Owner("gate") != "b" || bFlip.Flips() != 1 || bFlip.Progress("gate") != 0 {
		t.Fatalf("越过 -阈值应翻给 b，owner=%q flips=%d progress=%v",
			bFlip.Owner("gate"), bFlip.Flips(), bFlip.Progress("gate"))
	}

	exact := world()
	exact.Tick("gate", 1, 0, 10000)
	if exact.Owner("gate") != "" || exact.Flips() != 0 || exact.Progress("gate") != 100 {
		t.Fatalf("恰好到阈值不翻转，owner=%q flips=%d progress=%v",
			exact.Owner("gate"), exact.Flips(), exact.Progress("gate"))
	}

	afterFlip := NewWorld([]Point{{ID: "gate", Owner: "a"}})
	afterFlip.Tick("gate", 0, 12, 1000)
	if afterFlip.Owner("gate") != "b" || afterFlip.Flips() != 1 {
		t.Fatalf("反向越过阈值应改归属，owner=%q flips=%d", afterFlip.Owner("gate"), afterFlip.Flips())
	}
}

func TestProgressAlwaysBounded(t *testing.T) {
	instance := world()
	for index := 0; index < 1000; index++ {
		instance.Tick("gate", 1000, 0, 999)
		instance.Tick("gate", 0, 0, 50)
		progress := instance.Progress("gate")
		if progress < -100 || progress > 100 {
			t.Fatalf("进度越界：%v", progress)
		}
		if instance.Flips() > 0 && progress != 0 {
			t.Fatalf("翻转当步进度必须为 0：%v", progress)
		}
	}
}

func TestRepeatedSequenceDeterministic(t *testing.T) {
	steps := [][3]float64{{5, 0, 700}, {0, 0, 3000}, {0, 9, 400}, {2, 8, 1500}, {0, 0, 900}}
	run := func() (float64, string, int) {
		instance := world()
		for _, step := range steps {
			instance.Tick("gate", int(step[0]), int(step[1]), step[2])
		}
		return instance.Progress("gate"), instance.Owner("gate"), instance.Flips()
	}
	firstProgress, firstOwner, firstFlips := run()
	for index := 0; index < 5; index++ {
		progress, owner, flips := run()
		if progress != firstProgress || owner != firstOwner || flips != firstFlips {
			t.Fatalf("同序列结果不稳定：%v/%q/%d vs %v/%q/%d",
				progress, owner, flips, firstProgress, firstOwner, firstFlips)
		}
	}
}

func TestSnapshotCostIndependentOfPointCount(t *testing.T) {
	points := make([]Point, 0, 5000)
	for index := 0; index < 5000; index++ {
		points = append(points, Point{ID: fmt.Sprintf("p-%d", index)})
	}
	instance := NewWorld(points)
	for query := 0; query < 10000; query++ {
		instance.Snapshot()
	}
	if scanned := instance.Scanned(); scanned > 3000 {
		t.Fatalf("查询不许全表扫描，Scanned 应不随点数增长，实际 %d", scanned)
	}

	snapshot := instance.Snapshot()
	if len(snapshot) != 5000 {
		t.Fatalf("快照应包含全部 5000 个点，实际 %d", len(snapshot))
	}
	first := snapshot
	second := instance.Snapshot()
	for index := range first {
		if first[index] != second[index] {
			t.Fatal("快照顺序应稳定")
		}
	}
}

func TestRandomizedDifferential200(t *testing.T) {
	source := rand.New(rand.NewSource(20240930))
	for group := 0; group < 200; group++ {
		actual := world()
		reference := &refState{}

		mode := group % 4
		for step := 0; step < 60; step++ {
			var aCount, bCount int
			switch mode {
			case 0: // 只有 a 推进，穿插无人争夺
				if source.Intn(3) != 0 {
					aCount = source.Intn(16)
				}
			case 1: // 只有 b 推进，穿插无人争夺
				if source.Intn(3) != 0 {
					bCount = source.Intn(16)
				}
			case 2: // 双向争夺
				aCount = source.Intn(20)
				bCount = source.Intn(20)
			case 3: // 强争夺 + 无人争夺，制造频繁翻转
				if source.Intn(2) == 0 {
					aCount = source.Intn(30)
				} else {
					bCount = source.Intn(30)
				}
			}
			dtMs := []float64{1, 50, 100, 333, 1000, 2500, 7000}[source.Intn(7)]

			actual.Tick("gate", aCount, bCount, dtMs)
			refTick(reference, aCount, bCount, dtMs)

			if math.Abs(actual.Progress("gate")-reference.progress) > 1e-9 {
				t.Fatalf("组 %d 步 %d 进度不一致：实际 %v 参考 %v（%d,%d,dt=%v）",
					group, step, actual.Progress("gate"), reference.progress, aCount, bCount, dtMs)
			}
			if actual.Owner("gate") != reference.owner {
				t.Fatalf("组 %d 步 %d 归属不一致：实际 %q 参考 %q",
					group, step, actual.Owner("gate"), reference.owner)
			}
			if actual.Flips() != reference.flips {
				t.Fatalf("组 %d 步 %d 翻转次数不一致：实际 %d 参考 %d",
					group, step, actual.Flips(), reference.flips)
			}
		}
	}
}

func BenchmarkScale(b *testing.B) {
	const pointCount = 5000
	points := make([]Point, 0, pointCount)
	for index := 0; index < pointCount; index++ {
		points = append(points, Point{ID: string(rune(index + 1))})
	}
	instance := NewWorld(points)
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if index%2 == 0 {
			instance.Tick(points[index%pointCount].ID, 3, 1, 10)
		} else {
			instance.Snapshot()
		}
	}
}
