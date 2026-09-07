package server

import (
	"context"
	"time"

	"crow/internal/biz"
)

type InjectWorker struct {
	uc *biz.InjectUsecase
}

func NewInjectWorker(uc *biz.InjectUsecase) *InjectWorker {
	return &InjectWorker{uc: uc}
}

func (w *InjectWorker) Start(ctx context.Context) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			_ = w.uc.DispatchOnce(ctx)
		}
	}
}

func (w *InjectWorker) Stop(context.Context) error {
	return nil
}
