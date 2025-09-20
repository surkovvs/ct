package zorro

import (
	"math/bits"
	"sync/atomic"
)

type Zorro struct {
	status *uint64
}

type (
	Status uint64
	Mask   uint64
)

func New() Zorro {
	var status uint64
	return Zorro{
		status: &status,
	}
}

func (z Zorro) GetStatus() Status {
	return Status(atomic.LoadUint64(z.status))
}

// SetStatus - concurrently safe setup bits.
func (z Zorro) SetStatus(status Status, mask Mask) {
	for {
		if z.TrySetStatus(status, mask) {
			return
		}
	}
}

func (z Zorro) TrySetStatus(status Status, mask Mask) bool {
	cur := atomic.LoadUint64(z.status)
	upd := Status(cur).SetWithMask(status, mask)
	return atomic.CompareAndSwapUint64(z.status, cur, upd)
}

// TryChangeStatus - compare and swap but masked.
func (z Zorro) TryChangeStatus(prev, next Status, mask Mask) bool {
	cur := atomic.LoadUint64(z.status)
	if !prev.CompareMasked(Status(cur), mask) {
		return false
	}
	upd := Status(cur).SetWithMask(next, mask)
	return atomic.CompareAndSwapUint64(z.status, cur, upd)
}

// Querying example: status 1010 mask 0011 result 0010.
func (s Status) Querying(m Mask) uint64 {
	return uint64(s) & uint64(m)
}

// MaskedOn example: status 1010 mask 0011 result 1011.
func (s Status) MaskedOn(m Mask) uint64 {
	return uint64(s) | uint64(m)
}

// MaskedOff example: status 1010 mask 0011 result 1000.
func (s Status) MaskedOff(m Mask) uint64 {
	return uint64(s) &^ uint64(m)
}

// SetWithMask example: status 1010 mask 0011 set 0101 result 1001.
func (s Status) SetWithMask(set Status, m Mask) uint64 {
	return s.MaskedOff(m) | set.Querying(m)
}

func (s Status) CompareMasked(is Status, m Mask) bool {
	return s.Querying(m) == is.Querying(m)
}

// ShiftTrailingZeros example: status 1000 mask 1100 result 0010.
func (s Status) ShiftTrailingZeros(m Mask) uint64 {
	return uint64(s >> bits.TrailingZeros64(uint64(m)))
}
