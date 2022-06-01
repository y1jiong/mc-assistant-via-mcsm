package main

import (
	"mc-whitelist-team-manager-cli/internal/common"
	"mc-whitelist-team-manager-cli/internal/data"
)

func main() {
	c := common.Config{}
	//_ = c.InitToFile()
	_ = c.LoadFromFile()
	f := data.TeamsFormat{}
	_ = common.MarshalAndSave(f, c.DataFileName)
	_ = f.ExecuteWhiteTeamCommand(c.ApiUrl, c.ApiKey, c.ServerName)
}
