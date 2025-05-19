package client

import (
	"ATowerDefense/game"
	"fmt"
	"time"
)

func visualize(gm *game.Game) {
	for range gm.GC.FieldHeight {
		for range gm.GC.FieldWidth {
			fmt.Print("[]")
		}
		fmt.Println("")
	}
}

func Run(gc game.GameConfig) error {
	gm := game.NewGame(gc)
	if err := gm.Start(); err != nil {
		return err
	}

	for gm.GS.State != "stopped" {
		now := time.Now()

		gm.Iterate()

		visualize(gm)

		break

		time.Sleep(time.Second - time.Since(now))
	}

	return nil
}
