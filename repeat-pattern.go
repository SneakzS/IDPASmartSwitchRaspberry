package idpa

import "time"

const (
	WeekdayMonday     uint8 = 1 << time.Monday
	WeekdayTuesday    uint8 = 1 << time.Tuesday
	WeekdayWednessday uint8 = 1 << time.Wednesday
	WeekdayThursday   uint8 = 1 << time.Thursday
	WeekdayFriday     uint8 = 1 << time.Friday
	WeekdaySaturday   uint8 = 1 << time.Saturday
	WeekdaySunday     uint8 = 1 << time.Sunday
	WeekdayAll        uint8 = 127
)

const MonthAll uint16 = (1 << 13) - 1
const DayAll uint32 = (1 << 31) - 1

type RepeatPattern struct {
	MonthFlags   uint16
	DayFlags     uint32
	HourFlags    uint32
	MinuteFlags  uint64
	WeekdayFlags uint8
}

var monThue0800 = RepeatPattern{
	MonthFlags:   MonthAll,
	DayFlags:     DayAll,
	HourFlags:    (1 << 8),
	MinuteFlags:  1 << 0,
	WeekdayFlags: WeekdayMonday & WeekdayThursday,
}

func (rp RepeatPattern) Matches(t time.Time) bool {
	var (
		monthBit   uint16
		dayBit     uint32
		hourBit    uint32
		minuteBit  uint64
		weekdayBit uint8
	)

	monthBit = 1 << (t.Month() - 1)
	if rp.MonthFlags&monthBit == 0 {
		return false
	}

	dayBit = 1 << (t.Day() - 1)
	if rp.DayFlags&dayBit == 0 {
		return false
	}

	hourBit = 1 << t.Hour()
	if rp.HourFlags&hourBit == 0 {
		return false
	}

	minuteBit = 1 << t.Minute()
	if rp.MinuteFlags&minuteBit == 0 {
		return false
	}

	weekdayBit = 1 << t.Weekday()
	return rp.WeekdayFlags&weekdayBit != 0

}
