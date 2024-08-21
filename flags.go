package main

import (
	"strings"
	"time"
)

type FDuration time.Duration

func (d FDuration) String() string { return time.Duration(d).String() }
func (d *FDuration) Set(v string) error {
	td, err := time.ParseDuration(v)
	if err != nil {
		return err
	}
	*d = FDuration(td)
	return nil
}

type CSV []string

func (c CSV) String() string { return strings.Join(c, ",") }
func (c *CSV) Set(v string) error {
	*c = strings.Split(v, ",")
	return nil
}
