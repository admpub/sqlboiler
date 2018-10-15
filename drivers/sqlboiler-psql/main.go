package main

import (
	"github.com/admpub/sqlboiler/drivers"
	"github.com/admpub/sqlboiler/drivers/sqlboiler-psql/driver"
)

func main() {
	drivers.DriverMain(&driver.PostgresDriver{})
}
