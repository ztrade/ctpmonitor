package util

import "time"

type DayMinute int

type DayMinuteRange struct {
	Start DayMinute
	End   DayMinute
}

func NewDayMinute(hour, minute int) DayMinute {
	return DayMinute(hour*60 + minute)
}

type TradeTimeRange struct {
	Weekdays []time.Weekday
	Ranges   []DayMinuteRange
}

var (
	TradeTime = []DayMinuteRange{
		{Start: NewDayMinute(9, 0), End: NewDayMinute(11, 30)},
		{Start: NewDayMinute(13, 30), End: NewDayMinute(15, 00)},
		{Start: NewDayMinute(21, 00), End: NewDayMinute(26, 30)},
	}
)
