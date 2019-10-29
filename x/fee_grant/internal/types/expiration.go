package types

import "time"

// ExpiresAt is a point in time where something expires.
// It may be *either* block time or block height
type ExpiresAt struct {
	Time   time.Time `json:"time" yaml:"time"`
	Height int64     `json:"height" yaml:"height"`
}

// ExpiresAtTime creates an expiration at the given time
func ExpiresAtTime(t time.Time) ExpiresAt {
	return ExpiresAt{Time: t}
}

// ExpiresAtHeight creates an expiration at the given height
func ExpiresAtHeight(h int64) ExpiresAt {
	return ExpiresAt{Height: h}
}

// ValidateBasic performs basic sanity checks.
// Note that empty expiration is allowed
func (e ExpiresAt) ValidateBasic() error {
	if !e.Time.IsZero() && e.Height != 0 {
		return ErrInvalidDuration("both time and height are set")
	}
	if e.Height < 0 {
		return ErrInvalidDuration("negative height")
	}
	return nil
}

// IsZero returns true for an uninitialized struct
func (e ExpiresAt) IsZero() bool {
	return e.Time.IsZero() && e.Height == 0
}

// FastForward produces a new Expiration with the time or height set to the
// new value, depending on what was set on the original expiration
func (e ExpiresAt) FastForward(t time.Time, h int64) ExpiresAt {
	if !e.Time.IsZero() {
		return ExpiresAtTime(t)
	}
	return ExpiresAtHeight(h)
}

// IsExpired returns if the time or height is *equal to* or greater
// than the defined expiration point. Note that it is expired upon
// an exact match.
//
// Note a "zero" ExpiresAt is never expired
func (e ExpiresAt) IsExpired(t time.Time, h int64) bool {
	if !e.Time.IsZero() && !t.Before(e.Time) {
		return true
	}
	return e.Height != 0 && h >= e.Height
}

// IsCompatible returns true iff the two use the same units.
// If false, they cannot be added.
func (e ExpiresAt) IsCompatible(p Duration) bool {
	if !e.Time.IsZero() {
		return p.Clock > 0
	}
	return p.Block > 0
}

// Step will increase the expiration point by one Duration
// It returns an error if the Duration is incompatible
func (e ExpiresAt) Step(p Duration) (ExpiresAt, error) {
	if !e.IsCompatible(p) {
		return ExpiresAt{}, ErrInvalidDuration("expires_at and Duration have different units")
	}
	if !e.Time.IsZero() {
		e.Time = e.Time.Add(p.Clock)
	} else {
		e.Height += p.Block
	}
	return e, nil
}

// MustStep is like Step, but panics on error
func (e ExpiresAt) MustStep(p Duration) ExpiresAt {
	res, err := e.Step(p)
	if err != nil {
		panic(err)
	}
	return res
}

// PrepareForExport will deduct the dumpHeight from the expiration, so when this is
// reloaded after a hard fork, the actual number of allowed blocks is constant
func (e ExpiresAt) PrepareForExport(dumpTime time.Time, dumpHeight int64) ExpiresAt {
	if e.Height != 0 {
		e.Height -= dumpHeight
	}
	return e
}

// Duration is a repeating unit of either clock time or number of blocks.
// This is designed to be added to an ExpiresAt struct.
type Duration struct {
	Clock time.Duration `json:"clock" yaml:"clock"`
	Block int64         `json:"block" yaml:"block"`
}

// ClockDuration creates an Duration by clock time
func ClockDuration(d time.Duration) Duration {
	return Duration{Clock: d}
}

// BlockDuration creates an Duration by block height
func BlockDuration(h int64) Duration {
	return Duration{Block: h}
}

// ValidateBasic performs basic sanity checks
// Note that exactly one must be set and it must be positive
func (p Duration) ValidateBasic() error {
	if p.Block == 0 && p.Clock == 0 {
		return ErrInvalidDuration("neither time and height are set")
	}
	if p.Block != 0 && p.Clock != 0 {
		return ErrInvalidDuration("both time and height are set")
	}
	if p.Block < 0 {
		return ErrInvalidDuration("negative block step")
	}
	if p.Clock < 0 {
		return ErrInvalidDuration("negative clock step")
	}
	return nil
}
