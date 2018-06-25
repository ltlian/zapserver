// +build !solution

// Advanced Zap logger

package zlog

import (
	"sort"
	"sync"
	"time"

	zap "../"
)

// AdvancedZapLogger pairs ips and zaps, channels and stats, and channels and
// viewer counts
// TODO read/write specific locks for reduced blocking
type AdvancedZapLogger struct {
	ipMap   map[string]zap.ChZap
	stats   map[string]ZapStats
	chanMap ZapsMap
	mu      sync.Mutex
}

// NewAdvancedZapLogger creates an advanced zap logger
func NewAdvancedZapLogger() ZapLogger {
	advLogger := new(AdvancedZapLogger)
	// TODO consistent field types
	advLogger.chanMap = make(ZapsMap)
	advLogger.ipMap = make(map[string]zap.ChZap)
	advLogger.stats = make(map[string]ZapStats)
	return advLogger
}

// LogZap adds a zap to the log
func (azl *AdvancedZapLogger) LogZap(z zap.ChZap) {

	azl.mu.Lock()
	defer azl.mu.Unlock()

	// Increment viewer count for a given channel in the chanMap
	azl.chanMap[z.ToChan]++

	// IP exists in map; log duration and decrement view for the zap's FromChan
	if _, ok := azl.ipMap[z.IP]; ok {
		azl.logDuration(z)

		// Decrement viewer count based on the zap event's 'FromChan' channel
		azl.chanMap[z.FromChan]--

		// IP is not in map; add new IP key and zap value
	} else {
		azl.ipMap[z.IP] = z
	}
}

func (azl *AdvancedZapLogger) logDuration(z zap.ChZap) {
	dur := z.Duration(azl.ipMap[z.IP])

	// Ignore "flip-through" views
	if dur > minDur {
		if stats, ok := azl.stats[z.FromChan]; ok {
			// Channel exists in map; calculate a new average
			stats.SampleSize = stats.SampleSize + 1
			stats.AvgDur += (dur / time.Duration(stats.SampleSize))
			azl.stats[z.FromChan] = stats
		} else {
			// Channel does not exist in map; initialize new key and value
			// stats := (ZapStats{SampleSize: 1, AvgDur: dur})
			azl.stats[z.FromChan] = ZapStats{SampleSize: 1, AvgDur: dur}
		}

	}
}

// Entries returns the number of channels in the log set
func (azl *AdvancedZapLogger) Entries() int {
	if PrintTimes {
		defer zap.TimeElapsed(time.Now(), azl.String()+".Entries")
	}

	return len(azl.chanMap)
}

// String returns the name of the logger
func (azl *AdvancedZapLogger) String() string {
	return "Advanced Logger"
}

// Viewers returns the number of viewers for a given channel
func (azl *AdvancedZapLogger) Viewers(chName string) int {
	if PrintTimes {
		defer zap.TimeElapsed(time.Now(), azl.String()+".Viewers")
	}

	return azl.chanMap[chName]
}

// Channels returns a list of channels in the log
func (azl *AdvancedZapLogger) Channels() []string {
	if PrintTimes {
		defer zap.TimeElapsed(time.Now(), azl.String()+".Channels")
	}

	channels := make([]string, 0, len(azl.chanMap))

	for k := range azl.chanMap {
		channels = append(channels, k)
	}

	return channels
}

// ChannelsViewers creates a slice of ChannelViewers which is defined in zaplogger.go.
// This is the number of viewers for each channel.
// A positive non-zero input argument will return the given amount of elements, while a negative argument will return all elements in
// the list.
func (azl *AdvancedZapLogger) ChannelsViewers() []*ChannelViewers {
	var ChanViewersList []*ChannelViewers

	if PrintTimes {
		defer zap.TimeElapsed(time.Now(), azl.String()+".ChannelsViewers")
	}

	for key, value := range azl.chanMap { // This is a risky read if there is a concurrent write to the map
		chanViews := ChannelViewers{key, value}
		ChanViewersList = append(ChanViewersList, &chanViews)
	}

	return ChanViewersList
}

// FetchSorted returns a list of channels and viewers, sorted by viewers
// A positive non-zero input argument will return the given amount of elements, while a negative argument will return all elements in
// the list.
func (azl *AdvancedZapLogger) FetchSorted(i uint8) ChanViewersList {
	var bv ByViewers

	if PrintTimes {
		defer zap.TimeElapsed(time.Now(), azl.String()+".FetchSorted")
	}

	azl.mu.Lock()
	defer azl.mu.Unlock()

	bv.ChanViewersList = azl.ChannelsViewers()
	sort.Sort(sort.Reverse(ByViewers(bv)))

	if bv.ChanViewersList.Len() > int(i) {
		return bv.ChanViewersList[:i]
	}

	return bv.ChanViewersList
}

// FetchStats returns a map of channel and Zapstat pairs
// A positive non-zero input argument will return the given amount of elements, while a negative argument will return all elements in
// the list.
func (azl *AdvancedZapLogger) FetchStats() *map[string]ZapStats {
	if PrintTimes {
		defer zap.TimeElapsed(time.Now(), azl.String()+".FetchStats")
	}

	azl.mu.Lock()
	defer azl.mu.Unlock()

	return &azl.stats
}
