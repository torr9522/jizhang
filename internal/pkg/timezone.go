package pkg

import (
	"os"
	"sync"
	"time"
)

var (
	systemTimezone         *time.Location
	systemTimezoneOverride *time.Location
	systemTimezoneMu       sync.RWMutex
	systemTimezoneOnce     sync.Once
)

// GetSystemTimezone returns the system's default timezone.
// Priority: TZ environment variable > system local timezone.
// Result is cached after first call.
func GetSystemTimezone() *time.Location {
	systemTimezoneMu.RLock()
	override := systemTimezoneOverride
	systemTimezoneMu.RUnlock()
	if override != nil {
		return override
	}

	systemTimezoneOnce.Do(func() {
		// Try TZ environment variable first (Docker/container support)
		if tzEnv := os.Getenv("TZ"); tzEnv != "" {
			if loc, err := time.LoadLocation(tzEnv); err == nil {
				systemTimezone = loc
				return
			}
		}

		// Fall back to system local timezone
		systemTimezone = time.Local
	})
	return systemTimezone
}

// SetSystemTimezoneForTest overrides the process timezone until the returned
// restore function is called. Date-sensitive tests should pair this with a
// fixed clock so their calendar-day expectations do not depend on the host.
func SetSystemTimezoneForTest(loc *time.Location) func() {
	if loc == nil {
		loc = time.UTC
	}

	systemTimezoneMu.Lock()
	previous := systemTimezoneOverride
	systemTimezoneOverride = loc
	systemTimezoneMu.Unlock()

	return func() {
		systemTimezoneMu.Lock()
		systemTimezoneOverride = previous
		systemTimezoneMu.Unlock()
	}
}

// GetSystemTimezoneName returns the IANA timezone name (e.g., "Asia/Shanghai", "UTC").
func GetSystemTimezoneName() string {
	loc := GetSystemTimezone()
	// time.Local.String() returns "Local" which is not IANA format
	// We need to get the actual zone name
	if loc == time.UTC {
		return "UTC"
	}

	// Try to get zone name from current time
	now := NowIn(loc)
	zone, _ := now.Zone()

	// If zone is an abbreviation (e.g., "CST"), try to get IANA name from TZ env
	if tzEnv := os.Getenv("TZ"); tzEnv != "" {
		return tzEnv
	}

	// Last resort: return zone abbreviation or "Local"
	if zone != "" {
		return zone
	}
	return "Local"
}

// NormalizeDateInTimezone normalizes a time to start-of-day (00:00:00) in the given timezone.
func NormalizeDateInTimezone(t time.Time, loc *time.Location) time.Time {
	// Convert to target timezone first
	inZone := t.In(loc)
	// Create start-of-day in that timezone
	return time.Date(inZone.Year(), inZone.Month(), inZone.Day(), 0, 0, 0, 0, loc)
}

// TodayInTimezone returns the current date at 00:00:00 in the given timezone.
func TodayInTimezone(loc *time.Location) time.Time {
	return NormalizeDateInTimezone(Now(), loc)
}

// TodayInSystemTimezone returns the current system date at 00:00:00 in the configured system timezone.
func TodayInSystemTimezone() time.Time {
	return TodayInTimezone(GetSystemTimezone())
}

// DaysUntil calculates the number of days from now until the target date,
// both normalized to the given timezone.
func DaysUntil(target time.Time, loc *time.Location) int {
	return DaysUntilFrom(Now(), target, loc)
}

// DaysUntilFrom calculates the number of days from a reference time until the
// target date, both normalized to the given timezone.
func DaysUntilFrom(now, target time.Time, loc *time.Location) int {
	today := NormalizeDateInTimezone(now, loc)
	targetDate := NormalizeDateInTimezone(target, loc)
	diff := targetDate.Sub(today)
	return int(diff.Hours() / 24)
}
