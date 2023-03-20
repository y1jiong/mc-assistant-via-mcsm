package service

import (
	"math"
	"mc-assistant-via-mcsm/internal/common"
	"strconv"
)

const (
	mcDayTicks = 24000
)

func RunTicker(c common.Config, dayMinutes int) (err error) {
	dayMilliseconds := dayMinutes * 60 * 1000
	ticksPerDelay := float64(mcDayTicks) / float64(dayMilliseconds) * float64(c.DelayMilliseconds)
	timeAddCmd := "time add " + strconv.Itoa(int(math.Round(ticksPerDelay)))
	ticker := c.NewTicker()
	for {
		<-ticker.C
		err = c.SendCommand(timeAddCmd)
		if err != nil {
			return
		}
	}
}
