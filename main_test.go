package main

import (
	"testing"
	"time"
)

func TestParseName(t *testing.T) {
	c := func(isValid bool, name string, mustID string, mustTime time.Time) {
		valid, parsedID, parsedTime := ParseName(name)
		if valid == false && isValid == false {
			return
		}
		if parsedID != mustID {
			t.Errorf("filename '%s' parsedID must be '%s' but get '%s'", name, mustID, parsedID)
		}
		if !parsedTime.Equal(mustTime) {
			t.Errorf("filename '%s' time must be '%s' but get '%s'", name, mustTime, parsedTime)
		}
	}

	c(
		true,
		"vzdump-qemu-100-2019_03_28-23_59_59.vma",
		"100",
		time.Date(2019, 03, 28, 23, 59, 59, 0, time.UTC),
	)

	c(
		true,
		"vzdump-qemu-101-2019_01_05-00_00_00.vma.lzo",
		"101",
		time.Date(2019, 01, 05, 00, 00, 00, 0, time.UTC),
	)

	c(
		true,
		"vzdump-qemu-102-2019_12_02-20_00_02.vma.gz",
		"102",
		time.Date(2019, 12, 02, 20, 00, 02, 0, time.UTC),
	)

	// dump log file
	c(
		false,
		"vzdump-qemu-102-2019_12_02-20_00_02.log",
		"102",
		time.Time{},
	)

	// bad extension
	c(
		false,
		"vzdump-qemu-99999-2019_01_05-00_00_00",
		"99999",
		time.Date(2019, 01, 05, 00, 00, 00, 0, time.UTC),
	)

	// bad date
	c(
		false,
		"vzdump-qemu-99999-20_01_05-00_00_00",
		"99999",
		time.Time{},
	)

	// bad format
	c(
		false,
		"muzon.mp3",
		"99999",
		time.Time{},
	)

	// bad format
	c(
		false,
		"muzon.vma.mp3",
		"99999",
		time.Time{},
	)
}
