package models

import "time"

type Hutang struct {
	FakturHutang Faktur
	TokoDihutang Toko
	StockBarang  Stock
	ReturBarang  Retur
	SisaHutang   int
	JatuhTempo   time.Time
}
