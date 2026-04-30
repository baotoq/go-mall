package data

import (
	"context"
	"time"

	"gomall/app/catalog/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// Sweeper expires stale ACTIVE reservations on a fixed interval.
// Implements kratos transport.Server so it can be registered with kratos.Server().
type Sweeper struct {
	repo     biz.ReservationRepo
	interval time.Duration
	log      *log.Helper
	cancel   context.CancelFunc
}

func NewSweeperServer(repo biz.ReservationRepo, logger log.Logger) *Sweeper {
	return &Sweeper{
		repo:     repo,
		interval: 60 * time.Second,
		log:      log.NewHelper(logger),
	}
}

func (s *Sweeper) Start(ctx context.Context) error {
	ctx, s.cancel = context.WithCancel(ctx)
	go s.run(ctx)
	return nil
}

func (s *Sweeper) Stop(_ context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

func (s *Sweeper) run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := s.repo.ExpireStale(ctx)
			if err != nil {
				s.log.Errorf("sweeper: expire stale: %v", err)
			} else if n > 0 {
				s.log.Infof("sweeper: expired %d reservations", n)
			}
		}
	}
}
