package decoder

import "github.com/ozgur-yalcin/mfa/scan"

var DataMaskValues = []DataMask{
	{
		func(i, j int) bool {
			return ((i + j) & 0x01) == 0
		},
	},
	{
		func(i, j int) bool {
			return (i & 0x01) == 0
		},
	},
	{
		func(i, j int) bool {
			return j%3 == 0
		},
	},
	{
		func(i, j int) bool {
			return (i+j)%3 == 0
		},
	},
	{
		func(i, j int) bool {
			return (((i / 2) + (j / 3)) & 0x01) == 0
		},
	},
	{
		func(i, j int) bool {
			return (i*j)%6 == 0
		},
	},
	{
		func(i, j int) bool {
			return ((i * j) % 6) < 3
		},
	},
	{
		func(i, j int) bool {
			return ((i + j + ((i * j) % 3)) & 0x01) == 0
		},
	},
}

type DataMask struct {
	isMasked func(i, j int) bool
}

func (this DataMask) UnmaskBitMatrix(bits *scan.BitMatrix, dimension int) {
	for i := 0; i < dimension; i++ {
		for j := 0; j < dimension; j++ {
			if this.isMasked(i, j) {
				bits.Flip(j, i)
			}
		}
	}
}
