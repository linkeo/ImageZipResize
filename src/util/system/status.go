package system

import (
	"fmt"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"log"
	"os"
	"runtime"
	slices_ "slices"
)

var parallel uint64 = 4
var memoryLimit uint64 = 256 * 1024 * 1024

type ByteSize uint64

func (b ByteSize) String() string {
	const (
		KiB = 1024
		MiB = KiB * 1024
		GiB = MiB * 1024
	)

	switch {
	case b >= GiB:
		return fmt.Sprintf("%.2fGiB", float64(b)/GiB)
	case b >= MiB:
		return fmt.Sprintf("%.2fMiB", float64(b)/MiB)
	case b >= KiB:
		return fmt.Sprintf("%.2fKiB", float64(b)/KiB)
	default:
		return fmt.Sprintf("%dB", b)
	}
}

func init() {
	threads := GetCpuCores()
	if !allowParallel() {
		threads = 1
	}
	available := GetAvailableMemory() / 2
	memoryPerCore := available / threads
	if memoryPerCore < memoryLimit {
		parallel = max(1, available/memoryLimit)
	} else {
		memoryLimit = memoryPerCore
		parallel = threads
	}
}

func allowParallel() bool {
	envParallel, _ := os.LookupEnv("PARALLEL")
	if envParallel == "OFF" {
		log.Printf("run without parallel: PARALLEL=%s", envParallel)
		return false
	}
	if slices_.Contains(os.Args[1:], "--no-parallel") {
		log.Printf("run without parallel: --no-parallel")
		return false
	}
	return true
}

func GetParallelism() uint64 {
	return parallel
}

func GetMemoryLimit() ByteSize {
	return ByteSize(memoryLimit)
}

func GetCpuCores() uint64 {
	cnt, err := cpu.Counts(false)
	if err != nil {
		return uint64(runtime.NumCPU())
	}
	return uint64(cnt)
}

func GetAvailableMemory() uint64 {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return 4 * 1024 * 1024 * 1024
	}
	return vm.Available
}
