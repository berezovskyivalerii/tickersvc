package health

import "time"

type SysClock struct{}

func (SysClock) Now() time.Time { return time.Now() }
