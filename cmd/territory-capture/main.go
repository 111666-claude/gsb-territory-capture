// Command territory-capture 跑领地占领样例。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/territory"
)

func world() *territory.World {
	return territory.NewWorld([]territory.Point{{ID: "gate"}})
}

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("territory-capture", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sample := flags.String("sample", "time", "time / decay / flip / scan")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	switch *sample {
	case "time":
		one := world()
		one.Tick("gate", 1, 0, 1000)
		ten := world()
		ten.Tick("gate", 1, 0, 10000)
		fmt.Fprintf(stdout, "one=%.0f ten=%.0f\n", one.Progress("gate"), ten.Progress("gate"))
	case "decay":
		instance := world()
		instance.Tick("gate", 4, 0, 1000)
		for index := 0; index < 10; index++ {
			instance.Tick("gate", 0, 0, 1000)
		}
		fmt.Fprintf(stdout, "progress=%.0f\n", instance.Progress("gate"))
	case "flip":
		instance := world()
		instance.Tick("gate", 12, 0, 1000)
		fmt.Fprintf(stdout, "progress=%.0f owner=%s flips=%d\n",
			instance.Progress("gate"), instance.Owner("gate"), instance.Flips())
	case "scan":
		points := make([]territory.Point, 0, 3000)
		for index := 0; index < 3000; index++ {
			points = append(points, territory.Point{ID: fmt.Sprintf("p-%d", index)})
		}
		instance := territory.NewWorld(points)
		for query := 0; query < 3000; query++ {
			instance.Snapshot()
		}
		fmt.Fprintf(stdout, "scanned=%d\n", instance.Scanned())
	default:
		fmt.Fprintln(stderr, "需要 --sample time|decay|flip|scan")
		return 2
	}
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
