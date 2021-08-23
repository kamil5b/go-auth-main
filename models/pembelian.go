package models

type Pembelian struct {
	NomorFaktur        int
	Quantity           int
	TipeQuantity       string
	DiskontilPembelian int
	TotalHargaBeli     int
	TipePembayaran     string
	TokoPenjual        Toko
	StockBarang        Stock
}

/*
{
	nomorfaktur:
	quantity:
	tipequantity:
	diskontil:
	totalharga:
	tipepembayaran:
	toko:{
		nomor:
		nama:
		alamat:
	}
	stock:{
		nomorstock:
		barang:{
			kode:
			nama:
			tipebig:
			btm:
			tipemedium:
			mts:
			tipesmall:
			hargakecil:
			tipebarang:
		}
		expired:
		bigqty:
		medqty:
		smallqty:
		hargabeli:
	}
}
*/
