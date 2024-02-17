package models

import (
	"time"
)

type Revision struct {
	Version string    // version used in lookup
	Time    time.Time // commit time
}
