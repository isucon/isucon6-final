package scenario

import (
	"bytes"
)

// copied from react's src/renderers/shared/utils/adler32.js
// BSD License
//
// For React software
//
// Copyright (c) 2013-present, Facebook, Inc.
// All rights reserved.
func Adler32(s []byte) int32 {
	runes := bytes.Runes(s)
	mod := 65521
	a := 1
	b := 0
	i := 0
	l := len(runes)
	m := l & -4
	for i < m {
		n := 0
		if i+4096 > m {
			n = m
		} else {
			n = i + 4096
		}
		for ; i < n; i += 4 {
			a += int(runes[i])
			b += a
			a += int(runes[i+1])
			b += a
			a += int(runes[i+2])
			b += a
			a += int(runes[i+3])
			b += a
		}
		a %= mod
		b %= mod
	}
	for ; i < l; i++ {
		a += int(runes[i])
		b += a
	}
	a %= mod
	b %= mod
	return int32(a | b<<16)
}
