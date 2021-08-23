package controllers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber"
	"github.com/kamil5b/go-auth-main/database"
	"github.com/kamil5b/go-auth-main/models"
	"github.com/kamil5b/go-auth-main/utils"
)

type returquery struct {
	NomorRetur     int
	NomorFaktur    int
	Status         string
	NomorStock     int
	Quantity       int
	TipeQuantity   string
	DiskontilRetur int
	TotalNominal   int
	Description    string
}

//GET
func GetAllRetur(c *fiber.Ctx) error {
	var rets []returquery
	var returs []models.Retur

	database.DB.Table("retur").Find(&rets)
	for _, ret := range rets {
		var faktur models.Faktur
		database.DB.Table("faktur").Where("`NomorFaktur` = ?", ret.NomorFaktur).Find(&faktur)
		stock := GetStock(ret.NomorStock)

		retur := models.Retur{
			FakturBarang:   faktur,
			Status:         ret.Status,
			BarangRetur:    stock.BarangStock,
			Quantity:       ret.Quantity,
			TipeQuantity:   ret.TipeQuantity,
			DiskontilRetur: ret.DiskontilRetur,
			TotalNominal:   ret.TotalNominal,
			Description:    ret.Description,
		}
		returs = append(returs, retur)
	}

	return c.JSON(returs)
}

//POST
func ReturBarang(c *fiber.Ctx) error {
	/*
		{
			nomorfaktur:
			quantity:
			tipequantity:
			kodebarang:
			diskontil:
			desc:
		}
	*/
	var data map[string]string
	var faktur models.Faktur
	//var barang models.Barang
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	db := database.DB.Table("faktur").Where("`NomorFaktur` = ?", data["nomorfaktur"]).Scan(&faktur)
	if db.Error != nil {
		fmt.Println(db.Error)
		return c.JSON(fiber.Map{
			"message": "nomor faktur belum terdaftar",
		})
	}
	stock := GetStockByFakturBarang(faktur.NomorFaktur, data["kodebarang"])
	total := 0
	if stock.BarangStock.ConvertQty(data["tipequantity"], "small", dataint["quantity"]) != 0 {
		total = (stock.BarangStock.ConvertQty(
			data["tipequantity"],
			"small",
			dataint["quantity"],
		) * stock.HargaBeliKecil) - dataint["diskontil"]
	}
	retur := models.Retur{
		FakturBarang:   faktur,
		Status:         "RETUR",
		BarangRetur:    stock.BarangStock,
		Quantity:       dataint["quantity"],
		TipeQuantity:   data["tipequantity"],
		DiskontilRetur: dataint["diskontil"],
		TotalNominal:   total,
		Description:    data["desc"],
	}
	tipeqty := ""
	if strings.EqualFold(data["tipequantity"], retur.BarangRetur.TipeBigQty) {
		tipeqty = "BigQty"
	} else if strings.EqualFold(data["tipequantity"], retur.BarangRetur.TipeMediumQty) {
		tipeqty = "MediumQty"
	} else if strings.EqualFold(data["tipequantity"], retur.BarangRetur.TipeSmallQty) {
		tipeqty = "SmallQty"
	} else {
		return c.JSON(fiber.Map{
			"message": "tipe quantity tidak ditemukan pada barang",
		})
	}
	qty := 0
	query := ""
	if tipeqty == "BigQty" {
		qty = stock.BigQty - dataint["quantity"]
		query = "UPDATE `stock` SET BigQty = ? WHERE `stock`.`NomorStock` = ?"
	} else if tipeqty == "MediumQty" {
		qty = stock.MediumQty - dataint["quantity"]
		query = "UPDATE `stock` SET MediumQty = ? WHERE `stock`.`NomorStock` = ?"
	} else if tipeqty == "SmallQty" {
		qty = stock.SmallQty - dataint["quantity"]
		query = "UPDATE `stock` SET SmallQty = ? WHERE `stock`.`NomorStock` = ?"
	}
	if qty < 0 {
		qty = 0
	}
	retur.Quantity = qty
	database.DB.Exec(query, qty, stock.NomorStock)
	/*
		INSERT INTO retur(Status, NomorStock, Quantity, TipeQuantity,
			DiskontilRetur, TotalNominal, Description) VALUES (?,?,?,?,?,?,?)
	*/
	query = `INSERT INTO retur(NomorFaktur, Status, NomorStock, Quantity, TipeQuantity, 
		DiskontilRetur, TotalNominal, Description) VALUES (?,?,?,?,?,?,?,?)`
	database.DB.Exec(query,
		retur.FakturBarang.NomorFaktur,
		retur.Status,
		stock.NomorStock,
		qty,
		retur.TipeQuantity,
		retur.DiskontilRetur,
		retur.TotalNominal,
		retur.Description,
	)
	query = `UPDATE hutang SET NominalRetur = NominalRetur + ?, 
	Diskontil = Diskontil - ?, SisaHutang = SisaHutang - ? 
	WHERE NomorStock = ?`
	database.DB.Exec(query,
		retur.TotalNominal,
		retur.DiskontilRetur,
		retur.TotalNominal,
		stock.NomorStock,
	)
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func GetReturFaktur(nomorfaktur int) []models.Retur {

	var returs []models.Retur
	var r []returquery
	var faktur models.Faktur
	database.DB.Table("faktur").Where("`NomorFaktur` = ?", nomorfaktur).Find(&faktur)
	database.DB.Table("retur").Where("'NomorFaktur' = ?", nomorfaktur).Find(&r)
	for _, ret := range r {
		stock := GetStock(ret.NomorStock)
		retur := models.Retur{
			FakturBarang:   faktur,
			Status:         ret.Status,
			BarangRetur:    stock.BarangStock,
			Quantity:       ret.Quantity,
			TipeQuantity:   ret.TipeQuantity,
			DiskontilRetur: ret.DiskontilRetur,
			TotalNominal:   ret.TotalNominal,
			Description:    ret.Description,
		}
		returs = append(returs, retur)
	}
	return returs
}

func GetReturStock(nomorstock int) (models.Retur, error) {
	var retur models.Retur
	var faktur models.Faktur
	var r returquery
	r.NomorRetur = 0
	db := database.DB.Table("retur").Where("'NomorStock' = ?", nomorstock).Find(&r)
	if db.Error != nil {
		return models.Retur{}, db.Error
	}
	if r.NomorRetur == 0 {
		return models.Retur{}, errors.New("no retur")
	}
	database.DB.Table("faktur").Where("'NomorFaktur' = ?", r.NomorFaktur).Find(&faktur)
	stock := GetStock(r.NomorStock)
	retur = models.Retur{
		FakturBarang:   faktur,
		Status:         r.Status,
		BarangRetur:    stock.BarangStock,
		Quantity:       r.Quantity,
		TipeQuantity:   r.TipeQuantity,
		DiskontilRetur: r.DiskontilRetur,
		TotalNominal:   r.TotalNominal,
		Description:    r.Description,
	}
	return retur, nil
}
