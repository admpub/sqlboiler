package main

import (
	"github.com/admpub/sqlboiler/drivers"
	"github.com/admpub/sqlboiler/drivers/sqlboiler-mysql/driver"
)

func main() {
	drivers.DriverMain(&driver.MySQLDriver{})
}
