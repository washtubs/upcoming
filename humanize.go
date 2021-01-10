package upcoming

import (
	"fmt"
	"strings"
	"time"
)

var durationUnits []time.Duration = []time.Duration{
	time.Hour * 24 * 7,
	time.Hour * 24,
	time.Hour,
	time.Minute,
	time.Second,
}

var durationShort []string = []string{
	"w",
	"d",
	"h",
	"m",
	"s",
}

var durationLong []string = []string{
	"week",
	"day",
	"hour",
	"minute",
	"second",
}

var HumanizeDurationOpts = struct {
	NowThreshold time.Duration
	Short        bool
	Resolution   int
}{
	NowThreshold: time.Second * 5,
	Short:        true,
	Resolution:   2,
}

func HumanizeDuration(duration time.Duration) string {
	if duration < HumanizeDurationOpts.NowThreshold {
		return "now"
	}
	resolution := HumanizeDurationOpts.Resolution
	remaining := duration
	sb := strings.Builder{}
	delim := ""
	for i, unit := range durationUnits {
		if duration > unit {
			unitCount := remaining / unit
			remaining = remaining % unit
			if HumanizeDurationOpts.Short {
				sb.WriteString(fmt.Sprintf("%s%d%s", delim, unitCount, durationShort[i]))
			} else {
				pluralS := ""
				if unitCount > 1 {
					pluralS = "s"
				}
				sb.WriteString(fmt.Sprintf("%s%d %s%s", delim, unitCount, durationLong[i], pluralS))
			}
			delim = " "
			resolution = resolution - 1
		}
		if resolution == 0 {
			break
		}
	}

	return sb.String()
}
