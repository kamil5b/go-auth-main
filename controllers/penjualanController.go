package controllers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber"
	"github.com/kamil5b/go-auth-main/database"
	"github.com/kamil5b/go-auth-main/models"
	"github.com/kamil5b/go-auth-main/utils"
)

type sumpenjualan struct {
	NomorFaktur    int
	TanggalFaktur  time.Time
	TotalDiskontil int
	TotalPenjualan int
}
type summarypenjualan struct {
	Details        []sumpenjualan
	TotalDiskontil int
	TotalPenjualan int
}

//POST
func JualBarang(c *fiber.Ctx) error {
	/*
		{
			nomorfaktur:
			quantity:
			tipequantity:
			tipepembayaran:
			kodebarang:
			diskontil:
			nomorcustomer:
			jatuhtempo:
		}

		Alur :
		1. Toko udah diregister
		2. Barang udah diregister
		3. Faktur sudah dibuat
		3a. Nomor giro udah di register
		4. Buat stock dulu!
		4. Buat penjualan

	*/
	var data map[string]string
	var faktur models.Faktur
	var barang models.Barang
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	db := database.DB.Table("faktur").Where("`NomorFaktur` = ?", data["nomorfaktur"]).Scan(&faktur)
	if db.Error != nil {
		fmt.Println(db.Error)
		return c.JSON(fiber.Map{
			"message": "nomor faktur belum terdaftar",
		})
	}
	tipeqty := ""
	database.DB.Table("barang").Where("`KodeBarang` = ?", data["kodebarang"]).Find(&barang)
	if strings.EqualFold(data["tipequantity"], barang.TipeBigQty) {
		tipeqty = "BigQty"
	} else if strings.EqualFold(data["tipequantity"], barang.TipeMediumQty) {
		tipeqty = "MediumQty"
	} else if strings.EqualFold(data["tipequantity"], barang.TipeSmallQty) {
		tipeqty = "SmallQty"
	} else {
		return c.JSON(fiber.Map{
			"message": "tipe quantity tidak ditemukan pada barang",
		})
	}

	dataint := utils.MapStringToInt(data)

	stocks, err := GetStockByKode(data["kodebarang"])
	if err != nil {
		return c.JSON(fiber.Map{
			"message": "stock untuk barang ini habis",
		})
	}
	fmt.Println(stocks)
	process := func(stock models.Stock, qty int) error {
		totalhargajual := 0
		if stock.BarangStock.ConvertQty(data["tipequantity"],
			"small", dataint["quantity"]) != 0 {
			totalhargajual = (stock.BarangStock.ConvertQty(
				data["tipequantity"],
				stock.BarangStock.TipeSmallQty,
				qty,
			) * stock.BarangStock.HargaJualKecil) - dataint["diskontil"]
		} else {
			return c.JSON(fiber.Map{
				"message": "tipe quantity tidak ditemukan pada barang",
			})
		}

		query := `INSERT INTO penjualan(NomorStock, 
		Quantity, TipeQuantity, DiskontilPenjualan, 
		TotalHarga, TipePembayaran, NomorGiro, NomorCustomer) 
		VALUES (?,?,?,?,?,?,?,?)`
		if strings.ToLower(data["tipepembayaran"]) == "cash" {
			if dataint["nomorcustomer"] != 0 {
				db = database.DB.Exec(query,
					stock.NomorStock,
					qty,
					data["tipequantity"],
					dataint["diskontil"],
					totalhargajual,
					data["tipepembayaran"],
					nil,
					dataint["nomorcustomer"],
				)
			} else {
				db = database.DB.Exec(query,
					stock.NomorStock,
					qty,
					data["tipequantity"],
					dataint["diskontil"],
					totalhargajual,
					data["tipepembayaran"],
					nil,
					nil,
				)
			}
		} else if strings.ToLower(data["tipepembayaran"]) == "kredit" {
			if dataint["nomorcustomer"] == 0 {
				return c.JSON(fiber.Map{
					"message": "pembayaran kredit harus ada data customer",
				})
			}
			jatuhtempo, err := utils.ParsingDate(data["jatuhtempo"])
			if err != nil {
				return c.JSON(fiber.Map{
					"message": "tanggal jatuh tempo invalid",
				})
			}
			db = database.DB.Exec(query,
				stock.NomorStock,
				qty,
				data["tipequantity"],
				dataint["diskontil"],
				totalhargajual,
				data["tipepembayaran"],
				nil,
				dataint["nomorcustomer"],
			)
			PiutangBarang(dataint["nomorfaktur"], dataint["nomorcustomer"], stock.NomorStock, dataint["diskontil"], totalhargajual, jatuhtempo)
		} else {
			if dataint["nomorcustomer"] == 0 {
				return c.JSON(fiber.Map{
					"message": "pembayaran giro harus ada data customer",
				})
			}
			db = database.DB.Exec(query,
				stock.NomorStock,
				qty,
				data["tipequantity"],
				dataint["diskontil"],
				totalhargajual,
				"GIRO",
				data["tipepembayaran"],
				dataint["nomorcustomer"],
			)
		}
		if db.Error != nil {
			return c.JSON(fiber.Map{
				"message": "penjualan error 1",
			})
		}
		var notransaksi int
		query = "SELECT `TransaksiPenjualan` FROM penjualan ORDER BY `TransaksiPenjualan` DESC LIMIT 1"
		database.DB.Raw(query).Find(&notransaksi)
		query = "INSERT INTO `fakturpenjualan`(`NomorFaktur`, `TransaksiPenjualan`) VALUES (?,?)"
		db = database.DB.Exec(
			query,
			dataint["nomorfaktur"],
			notransaksi,
		)
		if db.Error != nil {
			query = "delete from  penjualan order by TransaksiPenjualan desc limit 1"
			database.DB.Exec(query)
			return c.JSON(fiber.Map{
				"message": "penjualan error 2",
			})
		}
		return nil
	}
	tmpqty := dataint["quantity"]
	stockprocess := func(bmsqty int, stock models.Stock) (int, error) {
		qty := 0
		if bmsqty >= tmpqty {
			fmt.Println("bmsqty >= tmpqty")
			qty = bmsqty - tmpqty
			err = process(stock, tmpqty)
			if err != nil {
				return qty, err
			}
			tmpqty = 0
		} else if bmsqty > 0 {
			//stock.BigQty < tmpqty
			fmt.Println("bmsqty > 0")
			tmpqty -= bmsqty
			err = process(stock, bmsqty)
			if err != nil {
				return qty, err
			}
		} else {
			fmt.Println("bmsqty == 0")
			return tmpqty, nil
		}
		return qty, nil
	}
	for _, stock := range stocks {

		qty := 0
		query := ""
		//fmt.Println(stock)
		fmt.Println("tmpqty:", tmpqty)
		if tipeqty == "BigQty" {
			qty, err = stockprocess(stock.BigQty, stock)
			if err != nil {
				return err
			}
			query = "UPDATE `stock` SET BigQty = ? WHERE `stock`.`NomorStock` = ?"
		} else if tipeqty == "MediumQty" {
			qty, err = stockprocess(stock.MediumQty, stock)
			if err != nil {
				return err
			}
			query = "UPDATE `stock` SET MediumQty = ? WHERE `stock`.`NomorStock` = ?"
		} else if tipeqty == "SmallQty" {
			qty, err = stockprocess(stock.SmallQty, stock)
			if err != nil {
				return err
			}
			query = "UPDATE `stock` SET SmallQty = ? WHERE `stock`.`NomorStock` = ?"
		}
		fmt.Println("qty:", qty)
		if qty != tmpqty {
			db = database.DB.Exec(query, qty, stock.NomorStock)
		}
		if db.Error != nil {
			fmt.Println(db.Error)
			return db.Error
		}
		if tmpqty == 0 {
			break
		}
	}
	if tmpqty != 0 {
		msg := "success with " + strconv.Itoa(dataint["quantity"]-tmpqty) + " data proceed"
		return c.JSON(fiber.Map{
			"message": msg,
		})
	}
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

//GET
func GetAllPenjualan(c *fiber.Ctx) error {
	type jual struct {
		NomorFaktur        int
		TanggalFaktur      time.Time
		KodeBarang         string
		NamaBarang         string
		Expired            time.Time
		Quantity           int
		TipeQuantity       string
		HargaJualKecil     int
		TipePembayaran     string
		DiskontilPenjualan int
		TotalHarga         int
		NomorCustomer      int
	}
	/*
		SELECT faktur.NomorFaktur,faktur.TanggalFaktur,
		stock.KodeBarang,barang.NamaBarang,stock.Expired,
		penjualan.Quantity,penjualan.TipeQuantity, barang.HargaJualKecil,
		penjualan.TipePembayaran, penjualan.DiskontilPenjualan, penjualan.TotalHarga,
		IF(`NomorCustomer` IS NULL ,0,`NomorCustomer`) AS NomorCustomer FROM penjualan
		JOIN fakturpenjualan on fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan
		JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur
		JOIN stock ON penjualan.NomorStock = stock.NomorStock
		JOIN barang ON barang.KodeBarang = stock.KodeBarang
	*/
	var penjualan []jual
	query := `SELECT faktur.NomorFaktur,faktur.TanggalFaktur, 
	stock.KodeBarang,barang.NamaBarang,stock.Expired, 
	penjualan.Quantity,penjualan.TipeQuantity, barang.HargaJualKecil,
	penjualan.TipePembayaran, penjualan.DiskontilPenjualan, penjualan.TotalHarga, 
	IF(NomorCustomer IS NULL ,0,NomorCustomer) AS NomorCustomer FROM penjualan 
	JOIN fakturpenjualan on fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan 
	JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur 
	JOIN stock ON penjualan.NomorStock = stock.NomorStock 
	JOIN barang ON barang.KodeBarang = stock.KodeBarang`
	db := database.DB.Raw(query).Find(&penjualan)
	if db.Error != nil {
		return db.Error
	}
	return c.JSON(penjualan)
}

//POST FAKTUR PEMBELIAN
func FakturPenjualan(c *fiber.Ctx) error {
	var data map[string]string
	/*
		"nomor" : ""
		"tanggal" : ""
	*/
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	nomor, err := strconv.Atoi(data["nomor"])
	if err != nil {
		return c.JSON(fiber.Map{
			"message": "error atoi nomor faktur",
		})
	}
	tanggal, err := utils.ParsingDate(data["tanggal"])
	if err != nil {
		return c.JSON(fiber.Map{
			"message": "error parsing date",
		})
	}
	/*
		faktur := models.Faktur{
			NomorFaktur:   nomor,
			TanggalFaktur: tanggal,
			TipeTransaksi: "PEMBELIAN",
		}
	*/
	query := `INSERT INTO faktur(NomorFaktur, TanggalFaktur, 
		TipeTransaksi) VALUES (?,?,?)`
	database.DB.Exec(
		query,
		nomor,
		tanggal,
		"PENJUALAN",
	)
	return c.JSON(fiber.Map{
		"message": "success",
	})
}

//GET FAKTUR
func GetFakturPenjualan(c *fiber.Ctx) error {
	/*
		"nomor" : ""
		"tanggal" : ""
	*/
	var faktur []models.Faktur
	database.DB.Table("faktur").Where("`TipeTransaksi` = \"PENJUALAN\"").Find(&faktur)

	return c.JSON(faktur)
}

//POST PENJUALAN PER TANGGAL
func PenjualanPerTanggal(c *fiber.Ctx) error {
	type jual struct {
		NomorFaktur        int
		TanggalFaktur      time.Time
		KodeBarang         string
		NamaBarang         string
		Expired            time.Time
		Quantity           int
		TipeQuantity       string
		HargaJualKecil     int
		TipePembayaran     string
		DiskontilPenjualan int
		TotalHarga         int
		NomorCustomer      int
	}
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}

	tanggalawal, err := utils.ParsingDate(data["tanggalawal"])
	if err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	tanggalakhir, err := utils.ParsingDate(data["tanggalakhir"])
	if err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	/*
		SELECT faktur.NomorFaktur,faktur.TanggalFaktur,
		stock.KodeBarang,barang.NamaBarang,stock.Expired,
		penjualan.Quantity,penjualan.TipeQuantity, barang.HargaJualKecil,
		penjualan.TipePembayaran, penjualan.DiskontilPenjualan, penjualan.TotalHarga,
		IF(`NomorCustomer` IS NULL ,0,`NomorCustomer`) AS NomorCustomer FROM penjualan
		JOIN fakturpenjualan on fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan
		JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur
		JOIN stock ON penjualan.NomorStock = stock.NomorStock
		JOIN barang ON barang.KodeBarang = stock.KodeBarang
	*/
	var penjualan []jual
	query := `SELECT faktur.NomorFaktur,faktur.TanggalFaktur, 
	stock.KodeBarang,barang.NamaBarang,stock.Expired, 
	penjualan.Quantity,penjualan.TipeQuantity, barang.HargaJualKecil,
	penjualan.TipePembayaran, penjualan.DiskontilPenjualan, penjualan.TotalHarga, 
	IF(NomorCustomer IS NULL ,0,NomorCustomer) AS NomorCustomer FROM penjualan 
	JOIN fakturpenjualan on fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan 
	JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur 
	JOIN stock ON penjualan.NomorStock = stock.NomorStock 
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	WHERE (faktur.TanggalFaktur BETWEEN ? AND ?) `
	db := database.DB.Raw(query, tanggalawal, tanggalakhir).Find(&penjualan)
	if db.Error != nil {
		return db.Error
	}
	return c.JSON(penjualan)
}

//POST PENJUALAN PER FAKTUR
func PenjualanPerFaktur(c *fiber.Ctx) error {
	type jual struct {
		NomorFaktur        int
		TanggalFaktur      time.Time
		KodeBarang         string
		NamaBarang         string
		Expired            time.Time
		Quantity           int
		TipeQuantity       string
		HargaJualKecil     int
		TipePembayaran     string
		DiskontilPenjualan int
		TotalHarga         int
		NomorCustomer      int
	}
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)
	/*
		SELECT faktur.NomorFaktur,faktur.TanggalFaktur,
		stock.KodeBarang,barang.NamaBarang,stock.Expired,
		penjualan.Quantity,penjualan.TipeQuantity, barang.HargaJualKecil,
		penjualan.TipePembayaran, penjualan.DiskontilPenjualan, penjualan.TotalHarga,
		IF(`NomorCustomer` IS NULL ,0,`NomorCustomer`) AS NomorCustomer FROM penjualan
		JOIN fakturpenjualan on fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan
		JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur
		JOIN stock ON penjualan.NomorStock = stock.NomorStock
		JOIN barang ON barang.KodeBarang = stock.KodeBarang
	*/
	var penjualan []jual
	query := `SELECT faktur.NomorFaktur,faktur.TanggalFaktur, 
	stock.KodeBarang,barang.NamaBarang,stock.Expired, 
	penjualan.Quantity,penjualan.TipeQuantity, barang.HargaJualKecil,
	penjualan.TipePembayaran, penjualan.DiskontilPenjualan, penjualan.TotalHarga, 
	IF(NomorCustomer IS NULL ,0,NomorCustomer) AS NomorCustomer FROM penjualan 
	JOIN fakturpenjualan on fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan 
	JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur 
	JOIN stock ON penjualan.NomorStock = stock.NomorStock 
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	WHERE faktur.NomorFaktur = ? AND faktur.TipeTransaksi = "PENJUALAN"`
	db := database.DB.Raw(query, dataint["nomorfaktur"]).Find(&penjualan)
	if db.Error != nil {
		return db.Error
	}
	return c.JSON(penjualan)
}

//POST PENJUALAN PER BARANG
func PenjualanPerBarang(c *fiber.Ctx) error {
	type jual struct {
		NomorFaktur        int
		TanggalFaktur      time.Time
		KodeBarang         string
		NamaBarang         string
		Expired            time.Time
		Quantity           int
		TipeQuantity       string
		HargaJualKecil     int
		TipePembayaran     string
		DiskontilPenjualan int
		TotalHarga         int
		NomorCustomer      int
	}
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	/*
		SELECT faktur.NomorFaktur,faktur.TanggalFaktur,
		stock.KodeBarang,barang.NamaBarang,stock.Expired,
		penjualan.Quantity,penjualan.TipeQuantity, barang.HargaJualKecil,
		penjualan.TipePembayaran, penjualan.DiskontilPenjualan, penjualan.TotalHarga,
		IF(`NomorCustomer` IS NULL ,0,`NomorCustomer`) AS NomorCustomer FROM penjualan
		JOIN fakturpenjualan on fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan
		JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur
		JOIN stock ON penjualan.NomorStock = stock.NomorStock
		JOIN barang ON barang.KodeBarang = stock.KodeBarang
	*/
	var penjualan []jual
	query := `SELECT faktur.NomorFaktur,faktur.TanggalFaktur, 
	stock.KodeBarang,barang.NamaBarang,stock.Expired, 
	penjualan.Quantity,penjualan.TipeQuantity, barang.HargaJualKecil,
	penjualan.TipePembayaran, penjualan.DiskontilPenjualan, penjualan.TotalHarga, 
	IF(NomorCustomer IS NULL ,0,NomorCustomer) AS NomorCustomer FROM penjualan 
	JOIN fakturpenjualan on fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan 
	JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur 
	JOIN stock ON penjualan.NomorStock = stock.NomorStock 
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	WHERE stock.KodeBarang = ?`
	db := database.DB.Raw(query, data["kodebarang"]).Find(&penjualan)
	if db.Error != nil {
		return db.Error
	}
	return c.JSON(penjualan)
}

//POST PENJUALAN TANGGAL/BARANG
func PenjualanPerTanggalBarang(c *fiber.Ctx) error {
	type jual struct {
		NomorFaktur        int
		TanggalFaktur      time.Time
		KodeBarang         string
		NamaBarang         string
		Expired            time.Time
		Quantity           int
		TipeQuantity       string
		HargaJualKecil     int
		TipePembayaran     string
		DiskontilPenjualan int
		TotalHarga         int
		NomorCustomer      int
	}
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}

	tanggalawal, err := utils.ParsingDate(data["tanggalawal"])
	if err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	tanggalakhir, err := utils.ParsingDate(data["tanggalakhir"])
	if err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	/*
		SELECT faktur.NomorFaktur,faktur.TanggalFaktur,
		stock.KodeBarang,barang.NamaBarang,stock.Expired,
		penjualan.Quantity,penjualan.TipeQuantity, barang.HargaJualKecil,
		penjualan.TipePembayaran, penjualan.DiskontilPenjualan, penjualan.TotalHarga,
		IF(`NomorCustomer` IS NULL ,0,`NomorCustomer`) AS NomorCustomer FROM penjualan
		JOIN fakturpenjualan on fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan
		JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur
		JOIN stock ON penjualan.NomorStock = stock.NomorStock
		JOIN barang ON barang.KodeBarang = stock.KodeBarang
	*/
	var penjualan []jual
	query := `SELECT faktur.NomorFaktur,faktur.TanggalFaktur, 
	stock.KodeBarang,barang.NamaBarang,stock.Expired, 
	penjualan.Quantity,penjualan.TipeQuantity, barang.HargaJualKecil,
	penjualan.TipePembayaran, penjualan.DiskontilPenjualan, penjualan.TotalHarga, 
	IF(NomorCustomer IS NULL ,0,NomorCustomer) AS NomorCustomer FROM penjualan 
	JOIN fakturpenjualan on fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan 
	JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur 
	JOIN stock ON penjualan.NomorStock = stock.NomorStock 
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	WHERE stock.KodeBarang = ? AND (faktur.TanggalFaktur BETWEEN ? AND ?) `
	db := database.DB.Raw(query, data["kodebarang"], tanggalawal, tanggalakhir).Find(&penjualan)
	if db.Error != nil {
		return db.Error
	}
	return c.JSON(penjualan)
}

//-----SUMMARY------

//GET SUMMARY
func SummaryPenjualan(c *fiber.Ctx) error {
	/*
		SELECT faktur.NomorFaktur,faktur.TanggalFaktur,
		SUM(penjualan.DiskontilPenjualan) AS TotalDiskontil,
		SUM(penjualan.TotalHarga) AS TotalPenjual
		FROM penjualan JOIN fakturpenjualan on
		fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan
		JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur
		JOIN stock ON penjualan.NomorStock = stock.NomorStock
		JOIN barang ON barang.KodeBarang = stock.KodeBarang
		GROUP BY faktur.NomorFaktur
	*/
	var sums []sumpenjualan
	query := `SELECT faktur.NomorFaktur,faktur.TanggalFaktur,
	SUM(penjualan.DiskontilPenjualan) AS TotalDiskontil,
	SUM(penjualan.TotalHarga) AS TotalPenjualan
	FROM penjualan JOIN fakturpenjualan on
	fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan
	JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur
	JOIN stock ON penjualan.NomorStock = stock.NomorStock
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	GROUP BY faktur.NomorFaktur`

	database.DB.Raw(query).Find(&sums)
	sum := summarypenjualan{
		Details:        sums,
		TotalDiskontil: 0,
		TotalPenjualan: 0,
	}
	for _, s := range sums {
		sum.TotalDiskontil += s.TotalDiskontil
		sum.TotalPenjualan += s.TotalPenjualan
	}
	return c.JSON(sum)
}

//POST SUMMARY PER FAKTUR
func SummaryPenjualanPerFaktur(c *fiber.Ctx) error {
	/*
		SELECT faktur.NomorFaktur,faktur.TanggalFaktur,
		SUM(penjualan.DiskontilPenjualan) AS TotalDiskontil,
		SUM(penjualan.TotalHarga) AS TotalPenjual
		FROM penjualan JOIN fakturpenjualan on
		fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan
		JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur
		JOIN stock ON penjualan.NomorStock = stock.NomorStock
		JOIN barang ON barang.KodeBarang = stock.KodeBarang
		GROUP BY faktur.NomorFaktur
	*/
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	dataint := utils.MapStringToInt(data)

	var sum sumpenjualan
	query := `SELECT faktur.NomorFaktur,faktur.TanggalFaktur,
	SUM(penjualan.DiskontilPenjualan) AS TotalDiskontil,
	SUM(penjualan.TotalHarga) AS TotalPenjualan
	FROM penjualan JOIN fakturpenjualan on
	fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan
	JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur
	JOIN stock ON penjualan.NomorStock = stock.NomorStock
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	WHERE faktur.NomorFaktur = ? AND faktur.TipeTransaksi = "PENJUALAN"`

	database.DB.Raw(query, dataint["nomorfaktur"]).Find(&sum)

	return c.JSON(sum)
}

//POST SUMMARY PER TANGGAL
func SummaryPenjualanPerTanggal(c *fiber.Ctx) error {
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}

	tanggalawal, err := utils.ParsingDate(data["tanggalawal"])
	if err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	tanggalakhir, err := utils.ParsingDate(data["tanggalakhir"])
	if err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	var sums []sumpenjualan
	query := `SELECT faktur.TanggalFaktur,
	SUM(penjualan.DiskontilPenjualan) AS TotalDiskontil,
	SUM(penjualan.TotalHarga) AS TotalPenjualan
	FROM penjualan JOIN fakturpenjualan on
	fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan
	JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur
	JOIN stock ON penjualan.NomorStock = stock.NomorStock
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	WHERE (faktur.TanggalFaktur BETWEEN ? AND ?)
	GROUP BY faktur.TanggalFaktur `

	database.DB.Raw(query, tanggalawal, tanggalakhir).Find(&sums)
	sum := summarypenjualan{
		Details:        sums,
		TotalDiskontil: 0,
		TotalPenjualan: 0,
	}
	for _, s := range sums {
		sum.TotalDiskontil += s.TotalDiskontil
		sum.TotalPenjualan += s.TotalPenjualan
	}
	return c.JSON(sum)
}

//GET SUMMARY PER BARANG
func SummaryPenjualanPerBarang(c *fiber.Ctx) error {
	type sumpenjualan struct {
		KodeBarang     string
		TotalDiskontil int
		TotalPenjualan int
	}
	type quantityjual struct {
		TotalQuantity int
		TipeQuantity  string
	}
	type penjualan struct {
		BarangJual     models.Barang
		TotalSmallQty  int
		TotalMediumQty int
		TotalBigQty    int
		TotalDiskontil int
		TotalPenjualan int
	}
	query := `SELECT barang.KodeBarang,
	SUM(penjualan.DiskontilPenjualan) AS TotalDiskontil,
	SUM(penjualan.TotalHarga) AS TotalPenjual FROM penjualan
	JOIN stock ON penjualan.NomorStock = stock.NomorStock
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	GROUP BY barang.KodeBarang`
	var sums []sumpenjualan
	var pens []penjualan
	database.DB.Raw(query).Find(&sums)
	for _, s := range sums {
		var q []quantityjual
		query = `SELECT SUM(penjualan.Quantity) AS TotalQuantity, 
		LOWER(penjualan.TipeQuantity) AS TipeQuantity FROM penjualan 
		JOIN stock ON penjualan.NomorStock = stock.NomorStock 
		JOIN barang ON barang.KodeBarang = stock.KodeBarang 
		WHERE barang.KodeBarang = ? GROUP BY LOWER(penjualan.TipeQuantity)`
		database.DB.Raw(query, s.KodeBarang).Find(&q)
		barang := GetBarang(s.KodeBarang)
		pen := penjualan{
			BarangJual:     barang,
			TotalSmallQty:  0,
			TotalMediumQty: 0,
			TotalBigQty:    0,
			TotalDiskontil: s.TotalDiskontil,
			TotalPenjualan: s.TotalPenjualan,
		}
		for _, x := range q {
			if strings.EqualFold(barang.TipeBigQty, x.TipeQuantity) {
				pen.TotalBigQty += x.TotalQuantity
			} else if strings.EqualFold(barang.TipeMediumQty, x.TipeQuantity) {
				pen.TotalMediumQty += x.TotalQuantity
			} else if strings.EqualFold(barang.TipeSmallQty, x.TipeQuantity) {
				pen.TotalSmallQty += x.TotalQuantity
			} else {
				return c.JSON(fiber.Map{
					"message": "tipe quantity invalid",
				})
			}
		}
		pens = append(pens, pen)
	}
	return c.JSON(pens)
}

//POST SUMMARY BARANG TANGGAL
func SummaryPenjualanPerBarangTanggal(c *fiber.Ctx) error {
	type sumpenjualan struct {
		KodeBarang     string
		TotalDiskontil int
		TotalPenjualan int
	}
	type quantityjual struct {
		TotalQuantity int
		TipeQuantity  string
	}
	type penjualan struct {
		BarangJual     models.Barang
		TotalSmallQty  int
		TotalMediumQty int
		TotalBigQty    int
		TotalDiskontil int
		TotalPenjualan int
	}
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	tanggalawal, err := utils.ParsingDate(data["tanggalawal"])
	if err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	tanggalakhir, err := utils.ParsingDate(data["tanggalakhir"])
	if err != nil {
		fmt.Println(err)
		fmt.Println(&data)
		return err
	}
	query := `SELECT barang.KodeBarang,
	SUM(penjualan.DiskontilPenjualan) AS TotalDiskontil,
	SUM(penjualan.TotalHarga) AS TotalPenjualan
	FROM penjualan JOIN fakturpenjualan on
	fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan
	JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur
	JOIN stock ON penjualan.NomorStock = stock.NomorStock
	JOIN barang ON barang.KodeBarang = stock.KodeBarang
	WHERE (faktur.TanggalFaktur BETWEEN ? AND ?)
	GROUP BY barang.KodeBarang`
	var sums []sumpenjualan
	var pens []penjualan
	database.DB.Raw(query, tanggalawal, tanggalakhir).Find(&sums)
	for _, s := range sums {
		var q []quantityjual
		query = `SELECT SUM(penjualan.Quantity) AS TotalQuantity, 
		LOWER(penjualan.TipeQuantity) AS TipeQuantity
		FROM penjualan JOIN fakturpenjualan on
		fakturpenjualan.TransaksiPenjualan=penjualan.TransaksiPenjualan
		JOIN faktur on faktur.NomorFaktur = fakturpenjualan.NomorFaktur
		JOIN stock ON penjualan.NomorStock = stock.NomorStock
		JOIN barang ON barang.KodeBarang = stock.KodeBarang
		WHERE (faktur.TanggalFaktur BETWEEN ? AND ?) AND barang.KodeBarang = ? 
		GROUP BY LOWER(penjualan.TipeQuantity)`
		database.DB.Raw(query, tanggalawal, tanggalakhir, s.KodeBarang).Find(&q)
		barang := GetBarang(s.KodeBarang)
		pen := penjualan{
			BarangJual:     barang,
			TotalSmallQty:  0,
			TotalMediumQty: 0,
			TotalBigQty:    0,
			TotalDiskontil: s.TotalDiskontil,
			TotalPenjualan: s.TotalPenjualan,
		}
		for _, x := range q {
			if strings.EqualFold(barang.TipeBigQty, x.TipeQuantity) {
				pen.TotalBigQty += x.TotalQuantity
			} else if strings.EqualFold(barang.TipeMediumQty, x.TipeQuantity) {
				pen.TotalMediumQty += x.TotalQuantity
			} else if strings.EqualFold(barang.TipeSmallQty, x.TipeQuantity) {
				pen.TotalSmallQty += x.TotalQuantity
			} else {
				return c.JSON(fiber.Map{
					"message": "tipe quantity invalid",
				})
			}
		}
		pens = append(pens, pen)
	}
	if pens == nil {
		return c.JSON(fiber.Map{
			"message": "laporan keuangan tidak ditemukan",
		})
	}
	return c.JSON(pens)
}
