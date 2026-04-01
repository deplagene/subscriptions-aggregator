package main

import (
	"context"

	"github.com/theartofdevel/logging"
)

func main() {
	ctx := context.Background()

	logger := logging.NewLogger(
		logging.WithLevel("info"),
		logging.WithIsJSON(true),
	)

	ctx = logging.ContextWithLogger(ctx, logger)
}
