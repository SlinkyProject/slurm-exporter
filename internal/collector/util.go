// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package collector

import (
	"math"

	"k8s.io/utils/ptr"

	api "github.com/SlinkyProject/slurm-client/api/v0043"
)

func ParseUint64NoVal(noVal *api.V0043Uint64NoValStruct) uint64 {
	if noVal == nil {
		return 0
	}
	if isSet := ptr.Deref(noVal.Set, false); !isSet {
		return 0
	}
	if isInfinite := ptr.Deref(noVal.Infinite, false); isInfinite {
		return math.MaxUint64
	}
	number := uint64(ptr.Deref(noVal.Number, 0))
	return number
}

func ParseUint32NoVal(noVal *api.V0043Uint32NoValStruct) uint32 {
	if noVal == nil {
		return 0
	}
	if isSet := ptr.Deref(noVal.Set, false); !isSet {
		return 0
	}
	if isInfinite := ptr.Deref(noVal.Infinite, false); isInfinite {
		return math.MaxUint32
	}
	number := uint32(ptr.Deref(noVal.Number, 0))
	return number
}
