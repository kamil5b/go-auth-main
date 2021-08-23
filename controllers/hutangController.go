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
func GetHutang(c *fiber.Ctx) error {
	type gethutang struct {
		NomorHutang  int
		NomorFaktur  int
		NomorToko    int
		NomorStock   int
		NominalRetur int
		Diskontil    int
		SisaHutang   int
		JatuhTempo   time.Time
	}
	/*

		type Hutang struct {
			FakturHutang Faktur
			TokoDihutang Toko
			StockBarang  Stock
			ReturBarang  Retur
			SisaHutang   int
			JatuhTempo   time.Time
		}

	*/
	var hutangs []models.Hutang
	var htng []gethutang
	database.DB.Table("hutang").Where("`SisaHutang` > 0").Find(&htng)
	for _, tmp := range htng {
		var faktur models.Faktur
		var toko models.Toko
		var stock models.Stock
		var retur models.Retur
		database.DB.Table("faktur").Where("`NomorFaktur` = ?", tmp.NomorFaktur).Find(&faktur)
		database.DB.Table("toko").Where("`NomorToko` = ?", tmp.NomorToko).Find(&toko)
		stock = GetStock(tmp.NomorStock)
		retur, err := GetReturStock(tmp.NomorStock)
		if err != nil {
			retur = models.Retur{}
		}
		hutang := models.Hutang{
			FakturHutang: faktur,
			TokoDihutang: toko,
			StockBarang:  stock,
			ReturBarang:  retur,
			SisaHutang:   tmp.SisaHutang,
			JatuhTempo:   tmp.JatuhTempo,
		}
		hutangs = append(hutangs, hutang)
	}
	return c.JSON(hutangs)
}

func HutangBarang(nomorfaktur, nomortoko, nomorstock, diskontil, nominal int, jatuhtempo time.Time) {
	/*
		var toko models.Toko
		var faktur models.Faktur
		query := "SELECT `TransaksiPembelian` FROM pembelian WHERE `TransaksiPembelian` = ?"
		database.DB.Raw(query, transaksipenjualan).Scan(&notransaksi)
		retur := GetReturFaktur(nomorfaktur)
		database.DB.Table("faktur").Where("`NomorFaktur` = ?", nomorfaktur).Find(&faktur)
		database.DB.Table("toko").Where("`NomorToko` = ?", nomortoko).Find(&toko)
		/*
			hutang := models.Hutang{
				FakturHutang: faktur,
				TokoDihutang: toko,
				ReturBarang:  retur,
				SisaHutang:   nominal,
				JatuhTempo:   jatuhtempo,
			}
			/*
			INSERT INTO hutang(NomorFaktur, NomorToko, NominalRetur,
				Diskontil, SisaHutang, JatuhTempo) VALUES (?,?,?,?,?,?)

		database.DB.Table("fakturpembelian").Select("`DiskontilPembelian`").Joins("join pembelian on ").Where("`NomorToko` = ?", nomortoko).Find(&toko)
	*/
	query := `INSERT INTO hutang(NomorFaktur, NomorToko, NomorStock, 
		NominalRetur, Diskontil, SisaHutang, JatuhTempo) VALUES (?,?,?,?,?,?,?)`
	database.DB.Exec(query,
		nomorfaktur,
		nomortoko,
		nomorstock,
		0,
		diskontil,
		nominal,
		jatuhtempo,
	)
}

//PUT
func BayarHutang(c *fiber.Ctx) error {
	/*
		{
			nomorhutang:
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
	query := "UPDATE hutang SET `SisaHutang` = `SisaHutang` - ? WHERE `NomorHutang` = ?"
	database.DB.Exec(query,
		dataint["nomorhutang"],
		dataint["bayar"],
	)
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func HutangNaik(c *fiber.Ctx) error {
	/*
		{
			nomorhutang:
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
	query := "UPDATE hutang SET `SisaHutang` = `SisaHutang` + ? WHERE `NomorHutang` = ?"
	database.DB.Exec(query,
		dataint["nomorhutang"],
		dataint["naik"],
	)
	return c.JSON(fiber.Map{
		"message": "success",
	})
}
