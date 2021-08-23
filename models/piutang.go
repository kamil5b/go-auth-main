package models

import "time"

type Piutang struct {
	FakturPiutang   Faktur
	CustomerPiutang Customer
	StockBarang     Stock
	ReturBarang     Retur
	SisaPiutang     int
	JatuhTempo      time.Time
}
