package bingsearch

import (
    "errors"
    "golang.org/x/time/rate"
)

var ErrBlocked = errors.New("bing blockage")

var RateLimit = rate.NewLimiter(rate.Inf, 0)
