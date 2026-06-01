package main

import (
	"fmt"

	"github.com/maahdima/mwp/api/config"
	"github.com/maahdima/mwp/api/dataservice"
	"github.com/maahdima/mwp/api/dataservice/model"
)

func main() {
	db, _ := dataservice.ConnectDB(config.GetDBConfig())

	dbIface := model.Interface{
		InterfaceID: "*11",
		Name:        "home-vpn",
		PrivateKey:  "key",
		PublicKey:   "key",
		ListenPort:  "13230",
	}

	err := db.Save(&dbIface).Error
	fmt.Println("Result:", err)
}
