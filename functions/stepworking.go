package functions

import (
	"context"
	"fmt"
	"time"
)

func StepFunction(ctx context.Context) {
	steps := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	for _, value := range steps {
		select {
		case <-ctx.Done():
			fmt.Println("StepFunction is stopped")
			return
		default:
			fmt.Printf("Step #%d\n", value)
			time.Sleep(3 * time.Second)
		}
	}
}
