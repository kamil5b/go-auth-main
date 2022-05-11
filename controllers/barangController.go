package controllers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/kamil5b/go-auth-main/database"
	"github.com/kamil5b/go-auth-main/models"
	"github.com/kamil5b/go-auth-main/utils"
)

//POST
func BarangBaru(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	/*
		barang := models.Barang{
			KodeBarang:     data["kode"],
			NamaBarang:     data["nama"],
			TipeBigQty:     data["tipebig"],
			BigToMedium:    btm,
			TipeMediumQty:  data["tipemedium"],
			MediumToSmall:  mts,
			TipeSmallQty:   data["tipesmall"],
			HargaJualKecil: hargakecil,
			TipeBarang:     data["tipebarang"],
		}
	*/
	query := `INSERT INTO barang(KodeBarang, NamaBarang, 
		TipeBigQty, BigToMedium, TipeMediumQty,
		MediumToSmall, TipeSmallQty, HargaJualKecil, 
		TipeBarang) VALUES (?,?,?,?,?,?,?,?,?)`
	database.DB.Exec(query,
		data["kode"],
		data["nama"],
		data["tipebig"],
		dataint["btm"],
		data["tipemedium"],
		dataint["mts"],
		data["tipesmall"],
		dataint["hargakecil"],
		data["tipebarang"],
	)

	return c.JSON(fiber.Map{
		"message": "success",
	})
}

//GET
func GetAllBarang(c *fiber.Ctx) error {
	/*
		kode:
		nama:
		tipebig:
		btm:
		tipemedium:
		mts:
		tipesmall:
		hargakecil:
		tipebarang:
	*/
	var barang []models.Barang
	database.DB.Table("barang").Find(&barang)
	return c.JSON(barang)
}

//Post
func GetSatuBarang(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	/*
		kode:
		nama:
		tipebig:
		btm:
		tipemedium:
		mts:
		tipesmall:
		hargakecil:
		tipebarang:
	*/
	var barang models.Barang
	database.DB.Table("barang").Where("`KodeBarang` = ?", data["kodebarang"]).Find(&barang)
	return c.JSON(barang)
}

func GetBarang(kodebarang string) models.Barang {
	var barang models.Barang
	database.DB.Table("barang").Where("`KodeBarang` = ?", kodebarang).Find(&barang)
	return barang
}
