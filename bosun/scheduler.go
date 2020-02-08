package bosun

import (
	"time"

	"github.com/dustin/go-humanize"
	"github.com/robfig/cron/v3"
)

func (b *Bosun) RegisterSchedulers() {
	// Register Auto Create Index Patterns
	if b.autoIndexPattern.Enabled {
		id := b.scheduler.Schedule(b.autoIndexPattern.Schedule, cron.FuncJob(b.AutoIndexPattern))

		b.autoIndexPattern.entry = b.scheduler.Entry(id)

		next := b.autoIndexPattern.Schedule.Next(time.Now())
		b.logger.Infof("Scheduled Auto Index Pattern Creation. Next run at %s (%s)", next.String(), humanize.Time(next))
	}
}
