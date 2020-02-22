package gnomock

import (
	"context"
	"time"
)

type Options struct {
	Image string
	Tag   string
	Port  int

	HealthcheckInterval time.Duration
	Healthcheck         HealthcheckFunc
	Init                InitFunc

	Ctx          context.Context
	StartTimeout time.Duration
	WaitTimeout  time.Duration
}

type HealthcheckFunc func(string, string) error

func nopHealthcheck(string, string) error {
	return nil
}

type InitFunc func(*Container) error

func nopInit(*Container) error {
	return nil
}
