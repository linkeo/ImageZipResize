package concurrent

import "sync"

func ForEach[T any](slice []T, iterator func(value T), maxWorkers int) {
	workers := min(len(slice), maxWorkers)
	if workers <= 0 {
		return
	}
	queue := make(chan T, workers)
	wg := new(sync.WaitGroup)
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for value := range queue {
				iterator(value)
			}
		}()
	}
	for _, value := range slice {
		queue <- value
	}
	close(queue)
	wg.Wait()
}
