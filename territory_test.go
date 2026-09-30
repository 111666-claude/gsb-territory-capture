package territory

import "testing"

func world() *World {
	return NewWorld([]Point{{ID: "gate"}})
}

func TestStartsNeutral(t *testing.T) {
	if world().Owner("gate") != "" {
		t.Fatal("一开始不该有归属")
	}
}

func TestMoreAttackersAdvance(t *testing.T) {
	instance := world()
	instance.Tick("gate", 2, 0, 1000)
	if instance.Progress("gate") <= 0 {
		t.Fatalf("人多的一方应该推进进度：%v", instance.Progress("gate"))
	}
}

func TestUnknownPointIsIgnored(t *testing.T) {
	instance := world()
	instance.Tick("missing", 1, 0, 1000)
	if instance.Progress("missing") != 0 {
		t.Fatal("不存在的点不该有进度")
	}
}

func TestScannedStartsAtZero(t *testing.T) {
	if world().Scanned() != 0 {
		t.Fatal("初始扫描数应该是 0")
	}
}
