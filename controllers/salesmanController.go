package controllers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kamil5b/go-auth-main/database"
	"github.com/kamil5b/go-auth-main/models"
	"github.com/kamil5b/go-auth-main/utils"
)

type querysales struct {
	NomorSales      int
	NIK             string
	NomorFaktur     int
	TanggalFaktur   time.Time
	TotalPenjualan  int
	NominalInsentif int
}

type querysalesman struct {
	NIK           string
	TanggalTarget time.Time
	NominalTarget int
	Status        string
}
type Salesmans struct {
	Salesman      models.User
	TanggalTarget time.Time
	NominalTarget int
	Status        string
}

func GetSalesman(c *fiber.Ctx) error {
	var user []models.User
	database.DB.Table("users").Where("role = ?", "SALESMAN").Find(&user)
	return c.JSON(user)
}

func InsertSales(nik string, faktur models.Faktur, totalhargajual int) error {
	var qs querysalesman
	qs.NIK = ""
	query := `SELECT * FROM sales WHERE NIK = ? AND NomorFaktur = ?`
	db := database.DB.Raw(query,
		nik,
		faktur.NomorFaktur,
	).Find(&qs)
	if db.Error != nil {
		return db.Error
	}
	if qs.NIK == "" {
		query = `INSERT INTO sales(NIK, NomorFaktur, TanggalFaktur, 
			TotalPenjualan) VALUES (?,?,?,?)`
		db = database.DB.Exec(query,
			nik,
			faktur.NomorFaktur,
			faktur.TanggalFaktur,
			totalhargajual,
		)
	} else {
		query = `UPDATE sales SET 
		TotalPenjualan=TotalPenjualan + ? 
		WHERE NIK=? AND NomorFaktur=?`
		db = database.DB.Exec(query,
			totalhargajual,
			nik,
			faktur.NomorFaktur,
		)
	}
	return db.Error
}

//GET
func GetAllTargetSales(c *fiber.Ctx) error {
	var que []querysalesman
	var sales []Salesmans
	database.DB.Table("targetsales").Find(&que)
	for _, q := range que {
		var user models.User
		database.DB.Table("users").Where("`nik` = ?", q.NIK).Find(&user)
		sale := Salesmans{
			Salesman:      user,
			TanggalTarget: q.TanggalTarget,
			NominalTarget: q.NominalTarget,
			Status:        q.Status,
		}
		sales = append(sales, sale)
	}
	return c.JSON(sales)
}

//POST
func GetMyTargetSales(c *fiber.Ctx) error {
	var data map[string]string
	/*
		{
			nik :
		}
	*/
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	var que querysalesman
	database.DB.Table("targetsales").Where("`NIK` = ?", data["nik"]).Find(&que)
	var user models.User
	database.DB.Table("users").Where("`nik` = ?", que.NIK).Find(&user)
	sale := Salesmans{
		Salesman:      user,
		TanggalTarget: que.TanggalTarget,
		NominalTarget: que.NominalTarget,
		Status:        que.Status,
	}
	return c.JSON(sale)
}

//GET ONLY SUPERVISOR
func GetAllSales(c *fiber.Ctx) error {

	var que []querysales
	var sales []models.Sales
	database.DB.Raw("SELECT * FROM `sales`").Find(&que)
	for _, s := range que {
		var karyawan models.User
		var faktur models.Faktur
		database.DB.Table("users").Where("nik = ?", s.NIK).Find(&karyawan)
		database.DB.Table("faktur").Where("`NomorFaktur` = ?", s.NomorFaktur).Scan(&faktur)
		sale := models.Sales{
			NomorSales:      s.NomorSales,
			Karyawan:        karyawan,
			FakturPenjualan: faktur,
			TotalPenjualan:  s.TotalPenjualan,
			NominalInsentif: s.NominalInsentif,
		}
		sales = append(sales, sale)
	}
	return c.JSON(sales)
}

//PUT
func EditInsentif(c *fiber.Ctx) error {
	var data map[string]string
	/*
		{
			nomorsales:
			insentif: //TIDAK TURUN - SUDAH TURUN <TIDAK DIPAKE>
			nominal:
		}
	*/
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	query := "UPDATE `sales` SET `NominalInsentif` = ? WHERE `NomorSales` = ?"
	db := database.DB.Exec(query,
		dataint["nominal"],
		dataint["nomorsales"],
	)
	if db.Error != nil {
		return db.Error
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

//POST
func GetIndividualSales(c *fiber.Ctx) error {
	var data map[string]string
	type individu struct {
		Sales          []models.Sales
		TotalSales     int
		TotalPenjualan int
		TotalInsentif  int
	}
	/*
		{
			nik :
		}
	*/
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	var que []querysales
	var sales individu
	sales.TotalSales = 0
	sales.TotalPenjualan = 0
	sales.TotalInsentif = 0
	database.DB.Table("sales").Where("`nik` = ?",
		data["nik"],
	).Find(&que)
	for _, s := range que {
		var faktur models.Faktur
		database.DB.Table("faktur").Where("`NomorFaktur` = ?", s.NomorFaktur).Scan(&faktur)
		sales.TotalInsentif += s.NominalInsentif
		sale := models.Sales{
			NomorSales:      s.NomorSales,
			FakturPenjualan: faktur,
			TotalPenjualan:  s.TotalPenjualan,
			NominalInsentif: s.NominalInsentif,
		}
		sales.Sales = append(sales.Sales, sale)
		sales.TotalSales += s.TotalPenjualan
		sales.TotalPenjualan++
	}
	return c.JSON(sales)
}
