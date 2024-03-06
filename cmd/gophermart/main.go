package main

import (
	"context"

	"github.com/stsg/gophermart2/internal/logger"
	"github.com/stsg/gophermart2/internal/server"
	"github.com/stsg/gophermart2/internal/services/shutdowner"
)

func main() {
	logger.New()
	server.Run(context.Background())
	<-shutdowner.Get().ChShutdowned
}
