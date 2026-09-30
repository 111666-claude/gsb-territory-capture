package territory

import (
	"math/rand"
	"testing"
)

// TestTimeBasedAdvance 进度只取决于真实时间，与调用次数无关。
func TestTimeBasedAdvance(t *testing.T) {
	once := NewWorld([]Point{{ID: "p"}})
	once.Tick("p", 3, 1, 1000)

	split := NewWorld([]Point{{ID: "p"}})
	for index := 0; index < 10; index++ {
		split.Tick("p", 3, 1, 100)
	}
	if once.Progress("p") != split.Progress("p") {
		t.Fatalf("同样的总时长结果应相同：%v != %v", once.Progress("p"), split.Progress("p"))
	}
	want := float64(2) * RatePerSec
	if once.Progress("p") != want {
		t.Fatalf("净人数 2 打 1 秒应为 %v，得到 %v", want, once.Progress("p"))
	}

	oneSec := NewWorld([]Point{{ID: "p"}})
	oneSec.Tick("p", 1, 0, 1000)
	tenSec := NewWorld([]Point{{ID: "p"}})
	tenSec.Tick("p", 1, 0, 10000)
	if tenSec.Progress("p") != 10*oneSec.Progress("p") {
		t.Fatalf("10 秒增量应为 1 秒的 10 倍：%v vs %v", tenSec.Progress("p"), oneSec.Progress("p"))
	}
}

// TestZeroDurationAndEqualCounts 零时长不推进；人数相等（非零）也不推进。
func TestZeroDurationAndEqualCounts(t *testing.T) {
	instance := NewWorld([]Point{{ID: "p"}})
	instance.Tick("p", 5, 0, 0)
	if instance.Progress("p") != 0 {
		t.Fatalf("dt=0 不该推进：%v", instance.Progress("p"))
	}
	instance.Tick("p", 2, 2, 1000)
	if instance.Progress("p") != 0 {
		t.Fatalf("双方人数相等不该推进：%v", instance.Progress("p"))
	}
}

// TestDecayTowardZero 无人争夺时向 0 衰减，且不会越过 0。
func TestDecayTowardZero(t *testing.T) {
	instance := NewWorld([]Point{{ID: "p"}})
	instance.Tick("p", 4, 0, 1000)
	if got := instance.Progress("p"); got != 40 {
		t.Fatalf("准备阶段进度应为 40，得到 %v", got)
	}
	for index := 0; index < 10; index++ {
		instance.Tick("p", 0, 0, 1000)
	}
	if got := instance.Progress("p"); got != 20 {
		t.Fatalf("衰减 10 秒后应为 20，得到 %v", got)
	}
	for index := 0; index < 100; index++ {
		instance.Tick("p", 0, 0, 1000)
	}
	if got := instance.Progress("p"); got != 0 {
		t.Fatalf("衰减不能越过 0，得到 %v", got)
	}
	if instance.Flips() != 0 {
		t.Fatal("衰减不该计翻转")
	}

	negative := NewWorld([]Point{{ID: "n"}})
	negative.Tick("p", 0, 3, 500)
	negative.Tick("n", 0, 3, 500)
	if got := negative.Progress("n"); got != -15 {
		t.Fatalf("b 方推进应为 -15，得到 %v", got)
	}
	negative.Tick("n", 0, 0, 1000)
	if got := negative.Progress("n"); got != -13 {
		t.Fatalf("负进度应向 0 衰减到 -13，得到 %v", got)
	}
}

// TestFlipAtThreshold 严格超过阈值才翻转，翻转后清零、归属正确、Flips 加一。
func TestFlipAtThreshold(t *testing.T) {
	atThreshold := NewWorld([]Point{{ID: "p"}})
	atThreshold.Tick("p", 10, 0, 1000)
	if atThreshold.Progress("p") != 100 || atThreshold.Owner("p") != "" || atThreshold.Flips() != 0 {
		t.Fatalf("恰好到阈值不翻转：progress=%v owner=%q flips=%d",
			atThreshold.Progress("p"), atThreshold.Owner("p"), atThreshold.Flips())
	}

	toA := NewWorld([]Point{{ID: "p"}})
	toA.Tick("p", 12, 0, 1000)
	if toA.Progress("p") != 0 || toA.Owner("p") != "a" || toA.Flips() != 1 {
		t.Fatalf("越过 +阈值应翻给 a 并清零：progress=%v owner=%q flips=%d",
			toA.Progress("p"), toA.Owner("p"), toA.Flips())
	}

	toB := NewWorld([]Point{{ID: "p"}})
	toB.Tick("p", 0, 20, 750) // 净人数 -20、0.75 秒 => -150，越过 -100
	if toB.Progress("p") != 0 || toB.Owner("p") != "b" || toB.Flips() != 1 {
		t.Fatalf("越过 -阈值应翻给 b 并清零：progress=%v owner=%q flips=%d",
			toB.Progress("p"), toB.Owner("p"), toB.Flips())
	}

	backAndForth := NewWorld([]Point{{ID: "p"}})
	backAndForth.Tick("p", 12, 0, 1000)
	backAndForth.Tick("p", 0, 12, 1000)
	backAndForth.Tick("p", 12, 0, 1000)
	if backAndForth.Owner("p") != "a" || backAndForth.Flips() != 3 {
		t.Fatalf("来回翻转 3 次后应归 a：owner=%q flips=%d",
			backAndForth.Owner("p"), backAndForth.Flips())
	}
}

// TestProgressClamped 进度始终夹在 [-100, 100]。
func TestProgressClamped(t *testing.T) {
	instance := NewWorld([]Point{{ID: "p"}})
	instance.Tick("p", 9, 0, 1000) // 90
	instance.Tick("p", 2, 0, 1000) // +20 = 110，越过阈值翻转清零，不会停留越界
	if got := instance.Progress("p"); got != 0 {
		t.Fatalf("越过阈值应清零，得到 %v", got)
	}
	instance.Tick("p", 9, 0, 1000)
	instance.Tick("p", 2, 0, 500) // 90 + 10 = 100，恰好到阈值不翻转
	if got := instance.Progress("p"); got != 100 {
		t.Fatalf("应夹在 100：%v", got)
	}

	neg := NewWorld([]Point{{ID: "q"}})
	neg.Tick("q", 0, 9, 1000)
	neg.Tick("q", 0, 2, 500)
	if got := neg.Progress("q"); got != -100 {
		t.Fatalf("应夹在 -100：%v", got)
	}
}

// TestSnapshotDoesNotScan Snapshot 不扫描索引表：Scanned 不随点数/查询次数增长，
// 且返回内容正确、互不共享底层数组。
func TestSnapshotDoesNotScan(t *testing.T) {
	points := make([]Point, 0, 3000)
	for index := 0; index < 3000; index++ {
		points = append(points, Point{ID: pointID(index)})
	}
	instance := NewWorld(points)
	for query := 0; query < 3000; query++ {
		instance.Snapshot()
	}
	if scanned := instance.Scanned(); scanned > 3000 {
		t.Fatalf("3000 次快照扫描量必须 <= 3000，得到 %d", scanned)
	}

	small := NewWorld([]Point{{ID: "x"}, {ID: "y"}})
	small.Tick("y", 1, 0, 500)
	got := small.Snapshot()
	if len(got) != 2 || got[0].ID != "x" || got[1].ID != "y" || got[1].Progress != 5 {
		t.Fatalf("快照内容或顺序错误：%+v", got)
	}
	got[1].Progress = 999
	if small.Progress("y") == 999 {
		t.Fatal("快照必须是副本，修改不应写回世界")
	}
}

func pointID(index int) string {
	return "p-" + itoa(index)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}

// referencePoint 是独立的参考实现（按时间推进 + 衰减 + 严格阈值翻转）。
type referencePoint struct {
	owner    string
	progress float64
	flips    int
}

func referenceTick(state *referencePoint, aCount, bCount int, dtMs float64) {
	dt := dtMs / 1000.0
	if aCount == 0 && bCount == 0 {
		drop := DecayPerSec * dt
		if state.progress > 0 {
			state.progress -= drop
			if state.progress < 0 {
				state.progress = 0
			}
		} else if state.progress < 0 {
			state.progress += drop
			if state.progress > 0 {
				state.progress = 0
			}
		}
	} else {
		state.progress += float64(aCount-bCount) * RatePerSec * dt
	}
	switch {
	case state.progress > FlipThreshold:
		state.progress = 0
		state.owner = "a"
		state.flips++
	case state.progress < -FlipThreshold:
		state.progress = 0
		state.owner = "b"
		state.flips++
	case state.progress > 100:
		state.progress = 100
	case state.progress < -100:
		state.progress = -100
	}
}

// TestReferenceDifferential 随机 200 组「人数 + 时间步长」序列，
// 含无人争夺与双向争夺，逐步与参考实现比对进度、归属和翻转数。
func TestReferenceDifferential(t *testing.T) {
	rng := rand.New(rand.NewSource(20260930))
	for sequence := 0; sequence < 200; sequence++ {
		actual := NewWorld([]Point{{ID: "p"}})
		reference := &referencePoint{}
		steps := 1 + rng.Intn(120)
		for step := 0; step < steps; step++ {
			var aCount, bCount int
			switch rng.Intn(4) {
			case 0: // 无人争夺
				aCount, bCount = 0, 0
			case 1: // a 推进
				aCount, bCount = rng.Intn(16), 0
			case 2: // b 推进
				aCount, bCount = 0, rng.Intn(16)
			case 3: // 双方争夺
				aCount, bCount = rng.Intn(16), rng.Intn(16)
			}
			dtMs := float64(rng.Intn(3000)) + float64(rng.Intn(10))*0.1
			actual.Tick("p", aCount, bCount, dtMs)
			referenceTick(reference, aCount, bCount, dtMs)

			if actual.Progress("p") != reference.progress {
				t.Fatalf("序列 %d 步 %d（a=%d b=%d dt=%v）进度不一致：实际 %v 参考 %v",
					sequence, step, aCount, bCount, dtMs, actual.Progress("p"), reference.progress)
			}
			if actual.Owner("p") != reference.owner {
				t.Fatalf("序列 %d 步 %d 归属不一致：实际 %q 参考 %q",
					sequence, step, actual.Owner("p"), reference.owner)
			}
			if actual.Flips() != reference.flips {
				t.Fatalf("序列 %d 步 %d 翻转数不一致：实际 %d 参考 %d",
					sequence, step, actual.Flips(), reference.flips)
			}
			if p := actual.Progress("p"); p < -100 || p > 100 {
				t.Fatalf("序列 %d 步 %d 进度越界：%v", sequence, step, p)
			}
		}
	}
}

// TestDeterministic 同一串推进序列重复执行结果相同。
func TestDeterministic(t *testing.T) {
	run := func() (float64, string, int) {
		instance := NewWorld([]Point{{ID: "p"}})
		rng := rand.New(rand.NewSource(42))
		for step := 0; step < 200; step++ {
			instance.Tick("p", rng.Intn(12), rng.Intn(12), float64(rng.Intn(2500)))
		}
		return instance.Progress("p"), instance.Owner("p"), instance.Flips()
	}
	firstProgress, firstOwner, firstFlips := run()
	for repeat := 0; repeat < 3; repeat++ {
		progress, owner, flips := run()
		if progress != firstProgress || owner != firstOwner || flips != firstFlips {
			t.Fatalf("重复执行结果不稳定：(%v,%q,%d) vs (%v,%q,%d)",
				progress, owner, flips, firstProgress, firstOwner, firstFlips)
		}
	}
}

// BenchmarkMixedLoad 5000 个点的混合推进/查询负载。
func BenchmarkMixedLoad(b *testing.B) {
	const pointCount = 5000
	points := make([]Point, 0, pointCount)
	ids := make([]string, 0, pointCount)
	for index := 0; index < pointCount; index++ {
		id := pointID(index)
		points = append(points, Point{ID: id})
		ids = append(ids, id)
	}
	instance := NewWorld(points)
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		id := ids[iteration%pointCount]
		instance.Tick(id, (iteration*7)%13, (iteration*5)%13, float64(iteration%500))
		if iteration%10 == 0 {
			_ = instance.Snapshot()
		}
		_ = instance.Progress(id)
		_ = instance.Owner(id)
	}
}
