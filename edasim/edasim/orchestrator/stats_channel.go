package orchestrator

import (
	"context"
)

type key int
const statsChannelsKey key = 0

type StatsChannels struct {
	ChJobProcessed chan struct{}
	ChProcessedFilesWritten chan struct{}
	ChJobCompleted chan struct{}
	ChUploadSignaled chan struct{}
}

func SetStatsChannel(ctx context.Context) context.Context {
	return context.WithValue(ctx, statsChannelsKey, InitializeStatsChannels()) 	
}

func GetStatsChannel(ctx context.Context) *StatsChannels {
	return ctx.Value(statsChannelsKey).(*StatsChannels)
} 

func InitializeStatsChannels() *StatsChannels {
	return &StatsChannels {
		ChJobProcessed : make(chan struct{}),
		ChProcessedFilesWritten : make(chan struct{}),
		ChJobCompleted : make(chan struct{}),
		ChUploadSignaled : make(chan struct{}),
	}
}

func (s *StatsChannels) JobProcessed() {
	s.ChJobProcessed <-struct{}{}
}

func (s *StatsChannels) ProcessedFilesWritten() {
	s.ChProcessedFilesWritten <-struct{}{}
}

func (s *StatsChannels) JobCompleted() {
	s.ChJobCompleted <-struct{}{}
}

func (s *StatsChannels) UploadSignaled() {
	s.ChUploadSignaled <-struct{}{}
}