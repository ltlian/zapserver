package zlog

// Viewers logger

import (
	"sort"
	"time"

	zap "../"
)

// NewViewersZapLogger creates a map based zap logger
func NewViewersZapLogger() ZapLogger {
	zm := make(ZapsMap)
	return &zm
}

// LogZap adds a zap to the log
func (zm *ZapsMap) LogZap(z zap.ChZap) {
	(*zm)[z.ToChan]++
	if (*zm)[z.FromChan] > 0 {
		(*zm)[z.FromChan]--
	} else {
		(*zm)[z.FromChan] = 0
	}
}

// Entries returns the number of channels in the log set
func (zm *ZapsMap) Entries() int {
	return len(*zm)
}

// String returns the name of the logger
func (zm *ZapsMap) String() string {
	return "Viewers Logger"
}

// Viewers returns the viewer count for the given channel
func (zm *ZapsMap) Viewers(chName string) int {
	// TODO: implementer locking
	if PrintTimes {
		defer zap.TimeElapsed(time.Now(), zm.String()+".Viewers")
	}

	// Only need to access viewers of channel in map
	return (*zm)[chName]
}

// Channels returns a list of channels in the log
func (zm *ZapsMap) Channels() []string {
	if PrintTimes {
		defer zap.TimeElapsed(time.Now(), zm.String()+".Channels")
	}

	channels := make([]string, 0, len(*zm)) //Max size is len of map to be efficient
	for k := range *zm {
		channels = append(channels, k)
	}

	return channels
}

// ChannelsViewers creates a slice of ChannelViewers, which is defined in zaplogger.go.
// This is the number of viewers for each channel.
// A positive non-zero input argument will return the given amount of elements, while a negative argument will return all elements in
// the list.
func (zm *ZapsMap) ChannelsViewers() []*ChannelViewers {
	var ChanViewersList ChanViewersList

	if PrintTimes {
		defer zap.TimeElapsed(time.Now(), zm.String()+".ChannelsViewers")
	}

	for key, value := range *zm {
		chanViews := ChannelViewers{key, value}
		ChanViewersList = append(ChanViewersList, &chanViews)
	}

	return ChanViewersList
}

// FetchSorted returns a list of channels and viewers, sorted by viewers
func (zm *ZapsMap) FetchSorted(i uint8) ChanViewersList {
	var bv ByViewers

	if PrintTimes {
		defer zap.TimeElapsed(time.Now(), zm.String()+".FetchSorted")
	}

	bv.ChanViewersList = zm.ChannelsViewers()
	sort.Sort(sort.Reverse(ByViewers(bv)))
	return bv.ChanViewersList[:i]
}

// FetchStats is not supported by viewerslogger and will return a nil map
func (zm *ZapsMap) FetchStats() *map[string]ZapStats {
	return nil
}
