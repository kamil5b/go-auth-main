package controllers

import (
	"fmt"

	"github.com/gofiber/fiber"
	"github.com/kamil5b/go-auth-main/database"
	"github.com/kamil5b/go-auth-main/models"
	"github.com/kamil5b/go-auth-main/utils"
)

func GetBank(c *fiber.Ctx) error {
	var bank []models.Bank
	database.DB.Table("bank").Find(&bank)
	return c.JSON(bank)
}

func GetCustomer(c *fiber.Ctx) error {
	var customer []models.Customer
	database.DB.Table("customer").Find(&customer)
	return c.JSON(customer)
}

func PostCustomer(c *fiber.Ctx) error {
	var data map[string]string
	//var barang models.Barang
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	/*
		{
			nama:
			nomor:
			alamat:
		}
	*/
	query := "INSERT INTO customer(NamaCustomer, NomorHP, Alamat) VALUES (?,?,?)"
	database.DB.Exec(query, data["nama"], data["nomor"], data["alamat"])
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func UpdateCustomer(c *fiber.Ctx) error {
	var data map[string]string
	//var barang models.Barang
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	/*
		{
			nomorcustomer:
			nama:
			nomor:
			alamat:
		}
	*/
	query := "UPDATE customer SET NamaCustomer = ?, NomorHP = ?, Alamat = ? WHERE NomorUrut = ?"
	database.DB.Exec(query, data["nama"], data["nomor"], data["alamat"], dataint["nomorcustomer"])
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func GetToko(c *fiber.Ctx) error {
	var toko []models.Toko
	database.DB.Table("toko").Find(&toko)
	return c.JSON(toko)
}

func PostToko(c *fiber.Ctx) error {
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	/*
		{
			nama:
			nomor:
			alamat:
		}
	*/
	query := "INSERT INTO toko(NamaToko, NomorTelepon, Alamat) VALUES (?,?,?)"
	database.DB.Exec(query, data["nama"], data["nomor"], data["alamat"])
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func UpdateToko(c *fiber.Ctx) error {
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	/*
		{
			nomortoko:
			nama:
			nomor:
			alamat:
		}
	*/
	query := "UPDATE toko SET NamaToko = ?, NomorTelepon = ?, Alamat = ? WHERE NomorToko = ?"
	database.DB.Exec(query, data["nama"], data["nomor"], data["alamat"], dataint["nomortoko"])
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

func GetGiro(c *fiber.Ctx) error {
	var giros []models.Giro
	var nobank int
	rows, err := database.DB.Table("giro").Rows()
	if err != nil {
		return err
	}
	for rows.Next() {
		var giro models.Giro
		var bank models.Bank
		rows.Scan(
			&giro.NomorGiro,
			&giro.Nominal,
			&giro.TanggalGiro,
			nobank,
		)
		database.DB.Table("bank").Find(&bank)
		giro.BankGiro = bank
		giros = append(giros, giro)
	}

	return c.JSON(giros)
}

func PostGiro(c *fiber.Ctx) error {
	var data map[string]string
	//var barang models.Barang
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	/*
		{
			nomorgiro:
			nominal:
			tanggal:
			nomorbank:
		}
	*/
	query := "INSERT INTO giro(NomorGiro, Nominal, TanggalGiro, NomorBank) VALUES (?,?,?,?)"
	tanggal, err := utils.ParsingDate(data["tanggal"])
	if err != nil {
		return err
	}
	database.DB.Exec(query, data["nomorgiro"], data["nominal"], tanggal, dataint["nomorbank"])
	return c.JSON(fiber.Map{
		"message": "success",
	})
}
