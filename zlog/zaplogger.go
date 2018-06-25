// Package zlog provides an interface for the various loggers
package zlog

import (
	"fmt"
	"time"

	zap ".."
)

// For the purpose of statistics, ignore view durations below this value
const minDur = time.Second * 5

// PrintTimes is used to determine whether calculation times should be logged to console
var PrintTimes bool

// ZapLogger is the interface used by the various loggers
type ZapLogger interface {
	LogZap(z zap.ChZap)
	Entries() int
	Viewers(channelName string) int
	Channels() []string
	ChannelsViewers() []*ChannelViewers
	FetchSorted(uint8) ChanViewersList
	FetchStats() *map[string]ZapStats
}

// ZapsMap holds channel-viewercount pairs
type ZapsMap map[string]int

// ZapStats holds a per-channel, per-viewer average viewing duration and the sample size that the duration is based on.
type ZapStats struct {
	AvgDur     time.Duration
	SampleSize uint32
}

// ChannelViewers holds a Channel-Viewers pair
type ChannelViewers struct {
	Channel string
	Viewers int
}

func (cv ChannelViewers) String() string {
	return fmt.Sprintf("%v, %v", cv.Channel, cv.Viewers)
}

// ChanViewersList holds an array of ChannelViewers
type ChanViewersList []*ChannelViewers

// ByViewers holds a ChanViewersList and is used for implementing the sorting interface
type ByViewers struct{ ChanViewersList }

// Len is needed to fulfill the sorting interface
func (t ChanViewersList) Len() int { return len(t) }

// Swap is needed to fulfill the sorting interface
func (t ChanViewersList) Swap(i, j int) { t[i], t[j] = t[j], t[i] }

// Less is needed to fulfill the sorting interface
func (s ByViewers) Less(i, j int) bool {
	return s.ChanViewersList[i].Viewers < s.ChanViewersList[j].Viewers
}
