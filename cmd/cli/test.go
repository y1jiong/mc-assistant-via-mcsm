package main

import (
	"mc-whitelist-team-manager-cli/internal/common"
	"mc-whitelist-team-manager-cli/internal/data"
)

func main() {
	c := common.Config{}
	//_ = c.InitToFile()
	_ = c.LoadFromFile()
	f := data.FinalFormat{}
	_ = common.MarshalAndSave(f, c.DataFileName)
	_ = f.ExecuteCommand(c.ApiUrl, c.ApiKey, c.ServerName)
}
