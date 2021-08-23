package controllers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber"
	"github.com/kamil5b/go-auth-main/database"
	"github.com/kamil5b/go-auth-main/models"
	"github.com/kamil5b/go-auth-main/utils"
)

//GET
func GetPiutang(c *fiber.Ctx) error {
	type getpiutang struct {
		NomorPiutang  int
		NomorFaktur   int
		NomorCustomer int
		NomorStock    int
		NominalRetur  int
		Diskontil     int
		SisaPiutang   int
		JatuhTempo    time.Time
	}
	/*

		type Piutang struct {
			FakturPiutang Faktur
			TokoDipiutang Toko
			StockBarang  Stock
			ReturBarang  Retur
			SisaPiutang   int
			JatuhTempo   time.Time
		}

	*/
	var piutangs []models.Piutang
	var htng []getpiutang
	database.DB.Table("piutang").Where("`SisaPiutang` > 0").Find(&htng)
	for _, tmp := range htng {
		var faktur models.Faktur
		var customer models.Customer
		var stock models.Stock
		var retur models.Retur
		database.DB.Table("faktur").Where("`NomorFaktur` = ?", tmp.NomorFaktur).Find(&faktur)
		database.DB.Table("customer").Where("`NomorUrut` = ?", tmp.NomorCustomer).Find(&customer)
		stock = GetStock(tmp.NomorStock)
		retur, err := GetReturStock(tmp.NomorStock)
		if err != nil {
			retur = models.Retur{}
		}
		piutang := models.Piutang{
			FakturPiutang:   faktur,
			CustomerPiutang: customer,
			StockBarang:     stock,
			ReturBarang:     retur,
			SisaPiutang:     tmp.SisaPiutang,
			JatuhTempo:      tmp.JatuhTempo,
		}
		piutangs = append(piutangs, piutang)
	}
	return c.JSON(piutangs)
}

func PiutangBarang(nomorfaktur, nomorcustomer, nomorstock, diskontil, nominal int, jatuhtempo time.Time) {
	/*
		var toko models.Toko
		var faktur models.Faktur
		query := "SELECT `TransaksiPembelian` FROM pembelian WHERE `TransaksiPembelian` = ?"
		database.DB.Raw(query, transaksipenjualan).Scan(&notransaksi)
		retur := GetReturFaktur(nomorfaktur)
		database.DB.Table("faktur").Where("`NomorFaktur` = ?", nomorfaktur).Find(&faktur)
		database.DB.Table("toko").Where("`NomorCustomer` = ?", nomortoko).Find(&toko)
		/*
			piutang := models.Piutang{
				FakturPiutang: faktur,
				TokoDipiutang: toko,
				ReturBarang:  retur,
				SisaPiutang:   nominal,
				JatuhTempo:   jatuhtempo,
			}
			/*
			INSERT INTO piutang(NomorFaktur, NomorCustomer, NominalRetur,
				Diskontil, SisaPiutang, JatuhTempo) VALUES (?,?,?,?,?,?)

		database.DB.Table("fakturpembelian").Select("`DiskontilPembelian`").Joins("join pembelian on ").Where("`NomorCustomer` = ?", nomortoko).Find(&toko)
	*/
	query := `INSERT INTO piutang(NomorFaktur, NomorCustomer, NomorStock, 
		NominalRetur, Diskontil, SisaPiutang, JatuhTempo) VALUES (?,?,?,?,?,?,?)`
	database.DB.Exec(query,
		nomorfaktur,
		nomorcustomer,
		nomorstock,
		0,
		diskontil,
		nominal,
		jatuhtempo,
	)
}

//PUT
func BayarPiutang(c *fiber.Ctx) error {
	/*
		{
			nomorpiutang:
			bayar:
		}
	*/
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	query := "UPDATE piutang SET `SisaPiutang` = `SisaPiutang` - ? WHERE `NomorPiutang` = ?"
	database.DB.Exec(query,
		dataint["nomorpiutang"],
		dataint["bayar"],
	)
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func PiutangNaik(c *fiber.Ctx) error {
	/*
		{
			nomorpiutang:
			naik:
		}
	*/
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	query := "UPDATE piutang SET `SisaPiutang` = `SisaPiutang` + ? WHERE `NomorPiutang` = ?"
	database.DB.Exec(query,
		dataint["nomorpiutang"],
		dataint["naik"],
	)
	return c.JSON(fiber.Map{
		"message": "success",
	})
}
