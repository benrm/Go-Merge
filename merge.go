package merge

import (
	"context"
)

// MergeSlices takes any number of sorted slices and produces a slice of the same type with the contents of the those slices sorted. The arguments are:
// * A function that accepts two values of the same type held by the slices and returns true if the first value is less than the second
// * Any number of slices
func MergeSlices[V any, T ~[]V](lessF func(V, V) bool, slices ...T) T {
	var length int
	for i, _ := range slices {
		length += len(slices[i])
	}
	ret := make(T, 0, length)
	indices := make([]int, len(slices))
	for len(slices) > 0 {
		var least int = -1
		var i int
		for i < len(slices) {
			if indices[i] < len(slices[i]) {
				if least < 0 || lessF(slices[i][indices[i]], slices[least][indices[least]]) {
					least = i
				}
				i++
			} else {
				slices[i] = slices[len(slices)-1]
				slices = slices[:len(slices)-1]
				indices[i] = indices[len(indices)-1]
				indices = indices[:len(indices)-1]
			}
		}
		if least >= 0 {
			ret = append(ret, slices[least][indices[least]])
			indices[least]++
		}
	}
	return ret
}

// first is used when getting the first value from each channel. It contains
// the value received, the channel index, and whether the channel is still
// open.
type first[V any] struct {
	v  V
	i  int
	ok bool
}

// MergeChannels takes any number of channel types that produce sorted output
// and returns a channel of that type that receives all the channels' output,
// merges it, and forwards it to a returned channel. The arguments are:
// * A context.Context
// * A function that accepts two values of the type sent by the channels and
// returns whether the first value is less than the second
// * Any number of channels
func MergeChannels[V any, T ~chan V](ctx context.Context, lessF func(V, V) bool, channels ...T) T {
	ret := make(T)
	go func() {
		defer close(ret)
		// populate the list of first values asynchronously
		firstChan := make(chan first[V])
		latest := make([]V, len(channels))
		for i, _ := range channels {
			go func(i int) {
				select {
				case v, ok := <-channels[i]:
					firstChan <- first[V]{v: v, i: i, ok: ok}
				case <-ctx.Done():
				}
			}(i)
		}
		for i := 0; i < len(channels); i++ {
			select {
			case f := <-firstChan:
				if f.ok {
					latest[f.i] = f.v
				} else {
					channels[f.i] = nil
				}
			case <-ctx.Done():
				return
			}
		}
		for len(channels) > 0 {
			var least int = -1
			var i int
			for i < len(channels) {
				if channels[i] != nil {
					if least < 0 || lessF(latest[i], latest[least]) {
						least = i
					}
					i++
				} else {
					channels[i] = channels[len(channels)-1]
					channels = channels[:len(channels)-1]
					latest[i] = latest[len(latest)-1]
					latest = latest[:len(latest)-1]
				}
			}
			if least >= 0 {
				select {
				case ret <- latest[least]:
				case <-ctx.Done():
					return
				}
				select {
				case v, ok := <-channels[least]:
					if ok {
						latest[least] = v
					} else {
						channels[least] = nil
					}
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return ret
}
