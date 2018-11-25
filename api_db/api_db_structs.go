package api_db

// Status for a single day of the month
// We have the amount of downtime for this day so we can add that as a tooltip
type StatusDay struct {
	StatusOnline bool
	StatusOffline bool
	StatusUnknown bool
	AmountMinuteDown float64
	AmountMinuteUp float64
}

// Stats for a single month
// We store what weekday this months starts on
// Also have a list of each day in the month
type StatusMonth struct {
	Name     string
	DaysEmpty []int
	Days     []StatusDay
}

// Struct for returning homepage data
type HomepageData struct {
	StatusOnline bool
	StatusOffline bool
	StatusUnknown bool
	TimeAgoString string
	Months []StatusMonth
}
