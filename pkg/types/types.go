package types

import (
	"strconv"
	"time"
)

type DurationInSeconds time.Duration

func (d DurationInSeconds) Value() time.Duration {
	return time.Duration(d)
}

func (d *DurationInSeconds) SetValue(s string) error {
	if s == "" {
		return nil
	}

	sec, err := strconv.Atoi(s)
	if err != nil {
		return err
	}

	*d = DurationInSeconds(time.Duration(sec) * time.Second)

	return nil

}
