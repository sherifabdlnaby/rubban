package rubban

import (
	"context"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/robfig/cron/v3"
	"github.com/sherifabdlnaby/rubban/log"
)

type scheduler struct {
	scheduler  cron.Cron
	logger     log.Logger
	context    context.Context
	specParser cron.Parser
	tasks      []task
}

type task struct {
	Name  string
	Entry cron.Entry
}

func newScheduler(ctx context.Context, logger log.Logger) *scheduler {
	return &scheduler{
		scheduler:  *cron.New(),
		context:    ctx,
		logger:     logger,
		specParser: cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor),
		tasks:      make([]task, 0),
	}
}

func (s *scheduler) Start() {
	s.logger.Infof("Starting scheduler...")

	s.scheduler.Start()

	// Print Tasks Next Runtime
	for _, task := range s.tasks {
		next := task.Entry.Schedule.Next(time.Now())
		s.logger.Infof("Next %s run at %s (%s)", task.Name, next.String(), humanize.Time(next))
	}
}

func (s *scheduler) Stop() {
	ctx := s.scheduler.Stop()

	// Wait for Running Jobs to finish.
	select {
	case <-ctx.Done():
		break
	case <-time.After(500 * time.Millisecond):
		s.logger.Infof("Waiting for running jobs to finish...")
		<-ctx.Done()
	}
}

func (s *scheduler) Register(spec string, job Task) error {

	schedule, err := s.specParser.Parse(spec)
	if err != nil {
		return err
	}

	entry := s.scheduler.Schedule(schedule, cron.FuncJob(func() {
		// TODO handle PANIC to not let all app crash if a single task crash.
		s.logger.Infof("Running %s...", job.Name())
		startTime := time.Now()

		job.Run(s.context)

		next := schedule.Next(time.Now())
		s.logger.Infof("Finished %s. (took ≈ %dms)", job.Name(), time.Since(startTime).Milliseconds())
		s.logger.Infof("Next %s run at %s (%s)", job.Name(), next.String(), humanize.Time(next))
	}))

	s.tasks = append(s.tasks, task{
		Name:  job.Name(),
		Entry: s.scheduler.Entry(entry),
	})

	s.logger.Infof("Registered %s", job.Name())
	return nil
}
