package main

import (
	"github.com/admpub/sqlboiler/drivers"
	"github.com/admpub/sqlboiler/drivers/sqlboiler-mssql/driver"
)

func main() {
	drivers.DriverMain(&driver.MSSQLDriver{})
}
