// +build !solution

// Simple Zap logger

package zlog

import (
	"sort"
	"sync"
	"time"

	zap "../"
)

type Zaps []zap.ChZap

var mutex = &sync.Mutex{}

func NewSimpleZapLogger() ZapLogger {
	zs := make(Zaps, 0)
	return &zs
}

// LogZap adds a zap to the log
func (zs *Zaps) LogZap(z zap.ChZap) {
	*zs = append(*zs, z)
}

// Entries returns the number of logged zap events
func (zs *Zaps) Entries() int {
	return len(*zs)
}

// String returns the name of the logger
func (zs *Zaps) String() string {
	return "Simple Logger"
}

// Viewers returns the current number of viewers for a given channel.
func (zs *Zaps) Viewers(chName string) int {

	if PrintTimes {
		defer zap.TimeElapsed(time.Now(), "simple.Viewers")
	}

	mutex.Lock()
	defer mutex.Unlock()
	var viewers int
	for _, zap := range *zs {
		switch chName {
		case zap.ToChan:
			viewers++
		case zap.FromChan:
			if viewers > 0 {
				viewers--
			}
		}
	}
	return viewers
}

// Channels creates a slice of the channels found in the zaps (both to and from).
func (zs *Zaps) Channels() []string {
	var channels []string

	if PrintTimes {
		defer zap.TimeElapsed(time.Now(), zs.String()+".Channels")
	}

	chanMap := make(map[string]int) //Using maps to easily find unique channels in Zaps
	for _, v := range *zs {
		chanMap[v.ToChan] = 1
		chanMap[v.FromChan] = 1
	}
	for k := range chanMap {
		channels = append(channels, k)
	}
	return channels
}

// ChannelsViewers creates a slice of ChannelViewers, which is defined in zaplogger.go.
// This is the number of viewers for each channel.
// A positive non-zero input argument will return the given amount of elements, while a negative argument will return all elements in
// the list.
func (zs *Zaps) ChannelsViewers() []*ChannelViewers {
	var ChanViewersList ChanViewersList

	if PrintTimes {
		defer zap.TimeElapsed(time.Now(), zs.String()+".ChannelsViewers")
	}

	channels := zs.Channels()
	for _, channel := range channels {
		chanViews := ChannelViewers{channel, zs.Viewers(channel)}
		ChanViewersList = append(ChanViewersList, &chanViews)
	}

	return ChanViewersList
}

// FetchSorted returns a list of channels and viewers, sorted by viewers
// A positive non-zero input argument will return the given amount of elements, while a negative argument will return all elements in
// the list.
func (zs *Zaps) FetchSorted(i uint8) ChanViewersList {
	var bv ByViewers

	if PrintTimes {
		defer zap.TimeElapsed(time.Now(), zs.String()+".FetchSorted")
	}

	bv.ChanViewersList = zs.ChannelsViewers()
	sort.Sort(sort.Reverse(ByViewers(bv)))
	return bv.ChanViewersList[:i]
}

// FetchStats is not supported by simplelogger and will return a nil map
func (zs *Zaps) FetchStats() *map[string]ZapStats {
	return nil
}
