package lib

import "time"

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func NowUTC() time.Time {
	return time.Now().UTC()
}
