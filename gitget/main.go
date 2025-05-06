package gitget

import (
	"context"
	"fmt"
	"os"

	"github.com/goaux/signals"
	"github.com/goaux/stacktrace/v2"
)

func Main() {
	ctx := context.Background()
	err := signals.NotifyContext(ctx, GitGet.ExecuteContext, os.Interrupt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", stacktrace.Format(err))
		os.Exit(1)
	}
}
