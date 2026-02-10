package converter

import (
	"context"
	"sync"
)

// BatchProgress reports progress for a batch conversion.
type BatchProgress struct {
	FileIndex  int
	TotalFiles int
	Filename   string
	Percent    float64
	Done       bool
	Err        error
}

// BatchConvert converts multiple files concurrently with a semaphore.
func BatchConvert(ctx context.Context, chdmanPath string, inputs []string, concurrency int, progressCh chan<- BatchProgress) []ConvertResult {
	if concurrency < 1 {
		concurrency = 1
	}

	results := make([]ConvertResult, len(inputs))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, input := range inputs {
		wg.Add(1)
		go func(idx int, inputPath string) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			output := OutputPath(inputPath)

			if progressCh != nil {
				progressCh <- BatchProgress{
					FileIndex:  idx,
					TotalFiles: len(inputs),
					Filename:   inputPath,
					Percent:    0,
				}
			}

			err := Convert(ctx, chdmanPath, inputPath, output, func(pct float64) {
				if progressCh != nil {
					progressCh <- BatchProgress{
						FileIndex:  idx,
						TotalFiles: len(inputs),
						Filename:   inputPath,
						Percent:    pct,
					}
				}
			})

			results[idx] = ConvertResult{
				InputPath:  inputPath,
				OutputPath: output,
				Err:        err,
			}

			if progressCh != nil {
				progressCh <- BatchProgress{
					FileIndex:  idx,
					TotalFiles: len(inputs),
					Filename:   inputPath,
					Percent:    100,
					Done:       true,
					Err:        err,
				}
			}
		}(i, input)
	}

	wg.Wait()
	if progressCh != nil {
		close(progressCh)
	}
	return results
}
