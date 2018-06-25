// +build !solution

package lab7

import (
	"fmt"
	"strings"
	"time"
)

const timeOnly = "15:04:05"
const datetimeFormat = "2006/01/02, 15:04:05"
const dateFormat = "2006/01/02"
const timeLen = len(datetimeFormat)

// ChZap represents a channel change event (Zap)
type ChZap struct {
	Time     time.Time
	IP       string
	ToChan   string
	FromChan string
}

// StatusChange represent a status change event
type StatusChange struct {
	Time   time.Time
	IP     string
	Status string
}

// NewSTBEvent takes a raw event string from the server and returns a ChZap or StatusChange event
func NewSTBEvent(event string) (*ChZap, *StatusChange, error) {

	if len(event) < 30 {
		return nil, nil, fmt.Errorf("NewSTBEvent: too short event string: %v", event)
	}

	fields := strings.Split(event[22:], ", ")
	if len(fields) < 2 {
		return nil, nil, fmt.Errorf("NewSTBEvent: event with too few fields: %v", event)
	}

	if len(fields) == 2 {
		// Parse event as status change (Currently only returns the raw string)
		ztat, err := parseStatus(fields)
		if err != nil {
			return nil, nil, err
		}

		return nil, ztat, nil
	}

	if len(fields) == 3 {
		// Parse event as channel change
		zap := parseZap(fields)

		parsedTime, err := time.Parse(datetimeFormat, event[:timeLen])
		if err != nil {
			err = fmt.Errorf("NewSTBEvent: failed to parse timestamp")
			return nil, nil, err
		}

		zap.Time = parsedTime
		return zap, nil, nil
	}

	return nil, nil, fmt.Errorf("NewSTBEvent: Could not parse event string: '%v'", event)
}

func (z ChZap) String() string {
	s := fmt.Sprintf("Time: %v, Date: %v, fromChan: %v, toChan: %v, IP: %v",
		z.Time, z.Date(), z.FromChan, z.ToChan, z.IP)
	return s
}

func (schg StatusChange) String() string {
	// StatusChange currently only holds the original event's string
	return schg.Status
}

// Duration returns the time between receiving (this) zap event and the provided event
func (z ChZap) Duration(provided ChZap) time.Duration {
	return z.Time.Sub(provided.Time)
}

func parseZap(event []string) *ChZap {
	var zap ChZap

	zap.IP = event[0]
	zap.FromChan = event[1]
	zap.ToChan = event[2]
	return &zap
}

func parseStatus(event []string) (*StatusChange, error) {
	// Currently only returns the raw string
	var ztat StatusChange
	ztat.Status = event[1]

	return &ztat, nil
}

// Date returns the date in string form
func (z ChZap) Date() string {
	return z.Time.Format(dateFormat)
}

type eventDateTimeError struct {
	err   error
	event string
}

func (e *eventDateTimeError) Error() string {
	return fmt.Sprintf("Could not parse timestamp in event '%v'\n%v", e.event, e.err.Error())
}

type eventLengthError struct {
	event string
}

func (e *eventLengthError) Error() string {
	return fmt.Sprintf("Event '%v' needs a minimum length of 30, but is of length %v", e.event, len(e.event))
}

type eventFieldsError struct {
	eventFields []string
}

func (e *eventFieldsError) Error() string {
	csv := strings.Join(e.eventFields, ", ")
	n := len(e.eventFields)
	return fmt.Sprintf("Event '%v' needs [2, 3] fields, but found %v", csv, n)
}
