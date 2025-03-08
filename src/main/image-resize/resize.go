package main

import (
	"ImageZipResize/tool/imagetool"
	"ImageZipResize/util"
	"ImageZipResize/util/concurrent"
	"ImageZipResize/util/fileutil"
	"ImageZipResize/util/filters"
	"ImageZipResize/util/slices"
	"ImageZipResize/util/system"
	"context"
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	slices_ "slices"
	"strings"
	"sync/atomic"
	"time"
)

var resizeTarget = image.Pt(1440, 1440)

type entry struct {
	root    string
	file    string
	isCover bool
	mem     system.ByteSize
}

func collect(files []string, dirs []string) ([]entry, map[string]struct{}) {
	entries := make([]entry, 0)
	roots := make(map[string]struct{})
	slices.ForEach(files, func(file string) {
		root := filepath.Dir(file)
		roots[root] = struct{}{}
		entries = append(entries, entry{root: root, file: file})
	})
	slices.ForEach(dirs, func(dir string) {
		root := filepath.Dir(dir)
		roots[root] = struct{}{}
		found, err := fileutil.ScanFiles(dir)
		if err != nil {
			log.Printf("scan files %s failed, %s", dir, err)
			return
		}
		slices.ForEach(found, func(file string) {
			entries = append(entries, entry{root: root, file: file})
		})
	})
	return entries, roots
}

func arrange(entries []entry) []entry {
	entries = slices.Filter(entries, func(e entry) bool {
		if !imagetool.IsSupportedImageFilename(e.file) {
			return false
		}
		if imagetool.IsOriginBackupPath(e.file) {
			return false
		}
		if imagetool.IsResizedPath(e.file) {
			return false
		}
		return true
	})
	slices_.SortStableFunc(entries, func(left entry, right entry) int {
		return strings.Compare(left.file, right.file)
	})
	for i := range entries {
		if i == 0 || filepath.Dir(entries[i].file) != filepath.Dir(entries[i-1].file) {
			entries[i].isCover = true
		}
	}
	return entries
}

var total int64 = 1
var totalWidth = 1
var todoAtomic = new(atomic.Int64)
var indexAtomic = new(atomic.Int64)
var doneAtomic = new(atomic.Int64)
var timeWindow *util.TimeWindow

func main() {
	args := slices.Filter(os.Args[1:], func(value string) bool {
		return !strings.HasPrefix(value, "--")
	})
	files := slices.Filter(args, filters.PathIsRegularFile)
	dirs := slices.Filter(args, filters.PathIsDirectory)
	for _, dir := range dirs {
		log.Printf("resize directory argument: %s", dir)
	}
	for _, file := range files {
		log.Printf("resize file argument: %s", file)
	}
	entries, roots := collect(files, dirs)
	entries = arrange(entries)

	total = int64(len(entries))
	totalWidth = len(fmt.Sprintf("%d", total))
	todoAtomic.Store(int64(total))
	indexAtomic.Store(0)
	doneAtomic.Store(0)

	par := system.GetParallelism()
	memoryLimit := system.GetMemoryLimit()
	timeWindow = util.NewTimeWindow(int(par*2), time.Second)

	if total == 0 {
		return
	}
	defer time.Sleep(time.Second)

	timeWindow.Append(time.Now())
	stopTitleLoop := startUpdateTitleLoop()
	defer stopTitleLoop()

	defer func() {
		for root := range roots {
			os.RemoveAll(fileutil.GetCacheDir(root))
		}
	}()

	failedEntries := make([]entry, 0)
	failedChan := make(chan entry, int(par))
	doneChan := make(chan struct{})
	go func() {
		for entry := range failedChan {
			failedEntries = append(failedEntries, entry)
		}
		close(doneChan)
	}()

	log.Printf("resize %d images with parallelism %d, each with %s memory limit.", total, par, memoryLimit)
	concurrent.ForEach(entries, func(en entry) {
		i := indexAtomic.Add(1)
		tag := fmt.Sprintf("%*d/%d", totalWidth, i, total)
		en.mem = memoryLimit
		if !resize(tag, en) {
			failedChan <- en
		}
	}, int(par))

	close(failedChan)
	<-doneChan

	failed := int64(len(failedEntries))
	if failed == 0 {
		return
	}

	memoryAvailable := system.GetMemoryAvailable()
	log.Printf("resize %d failed images sequentially, each with %s memory limit.", failed, memoryAvailable)
	timeWindow.Reset()
	timeWindow.SetDefaultAverage(10 * time.Second)
	timeWindow.Append(time.Now())
	todoAtomic.Store(failed)
	failedBase := total - failed + 1
	for i, en := range failedEntries {
		tag := fmt.Sprintf("%*d/%d", totalWidth, failedBase+int64(i), total)
		en.mem = memoryAvailable
		resize(tag, en)
	}
}

func resize(tag string, en entry) bool {
	result, err := imagetool.Resize(en.root, en.file, en.isCover, resizeTarget, imagetool.ModeContain.DoNotEnlarge(), en.mem)
	todoAtomic.Add(-1)
	if err != nil {
		log.Printf("[%s] %s resize failed, %s, %s, ETA %s", tag, resizeTarget, en.file, err, eta())
		return false
	}
	timeWindow.Append(time.Now())
	doneAtomic.Add(1)
	log.Printf("[%s] %s resize %7s, %s, ETA %s", tag, resizeTarget, compressRate(result), en.file, eta())
	return true
}

func compressRate(rate float64) string {
	deltaPercent := (rate - 1) * 100
	if deltaPercent == 0.0 {
		return "-0%"
	}
	if deltaPercent < 0.0 {
		return fmt.Sprintf("%.2f%%", deltaPercent)
	}
	return fmt.Sprintf("+%.2f%%", deltaPercent)
}

func eta() time.Duration {
	avg := timeWindow.Average()
	td := todoAtomic.Load()
	left := time.Duration(td) * avg
	return left.Truncate(time.Second)
}

func updateTitle() {
	nextTime := time.Now().Add(100 * time.Millisecond)
	done := doneAtomic.Load()
	title := fmt.Sprintf("[%d/%d] [ETA:%s] Image Resize", done, total, eta())
	percent := 100 * done / total
	fmt.Printf("\033]9;4;1;%d\a", percent)
	fmt.Printf("\033]0;%s\a", title)
	time.Sleep(nextTime.Sub(time.Now()))
}

func startUpdateTitleLoop() func() {
	ctx, cancel := context.WithCancel(context.Background())
	doneCh := make(chan struct{})
	go updateTitleLoop(ctx, doneCh)
	return func() {
		cancel()
		<-doneCh
	}
}

func updateTitleLoop(ctx context.Context, doneCh chan struct{}) {
	defer close(doneCh)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			updateTitle()
		}
	}
}
