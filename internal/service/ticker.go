package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"mc-assistant-via-mcsm/internal/common"
)

const minecraftDayTicks = 24_000

func RunTicker(ctx context.Context, config *common.Config, dayMinutes int) error {
	ticks, err := ticksPerInterval(dayMinutes, config.DelayDuration())
	if err != nil {
		return err
	}

	ticker := time.NewTicker(config.DelayDuration())
	defer ticker.Stop()
	timeAddCommand := "time add " + strconv.Itoa(ticks)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := config.SendCommand(ctx, timeAddCommand); err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				return err
			}
		}
	}
}

func ticksPerInterval(dayMinutes int, interval time.Duration) (int, error) {
	if dayMinutes <= 0 {
		return 0, errors.New("游戏日分钟数必须大于 0")
	}
	if interval <= 0 {
		return 0, errors.New("命令间隔必须大于 0")
	}

	dayNanoseconds := float64(dayMinutes) * float64(time.Minute)
	ticks := int(math.Round(float64(minecraftDayTicks) * float64(interval) / dayNanoseconds))
	if ticks < 1 {
		return 0, fmt.Errorf("命令间隔 %s 对 %d 分钟游戏日过短，无法增加至少 1 tick", interval, dayMinutes)
	}
	return ticks, nil
}
