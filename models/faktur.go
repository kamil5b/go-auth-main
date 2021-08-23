package models

import "time"

type Faktur struct {
	NomorFaktur   int
	TanggalFaktur time.Time
	TipeTransaksi string
}
