// Package load blank-imports every factory subdirectory so their init()
// registration runs. Any binary that wants ALL strategies loaded should
// import this single package; nothing else needed.
//
// Adding a new strategy? Run `go generate ./...` or manually add one
// `_ "github.com/neko233-com/game-ops/factory/<name>"` line below.
package load

import (
	_ "github.com/neko233-com/game-ops/factory/action_rpg"
	_ "github.com/neko233-com/game-ops/factory/adventure"
	_ "github.com/neko233-com/game-ops/factory/battle_royale"
	_ "github.com/neko233-com/game-ops/factory/card"
	_ "github.com/neko233-com/game-ops/factory/casual"
	_ "github.com/neko233-com/game-ops/factory/fighting"
	_ "github.com/neko233-com/game-ops/factory/fps"
	_ "github.com/neko233-com/game-ops/factory/gacha"
	_ "github.com/neko233-com/game-ops/factory/horror"
	_ "github.com/neko233-com/game-ops/factory/idle"
	_ "github.com/neko233-com/game-ops/factory/match3"
	_ "github.com/neko233-com/game-ops/factory/mmo"
	_ "github.com/neko233-com/game-ops/factory/mmorpg"
	_ "github.com/neko233-com/game-ops/factory/moba"
	_ "github.com/neko233-com/game-ops/factory/platformer"
	_ "github.com/neko233-com/game-ops/factory/puzzle"
	_ "github.com/neko233-com/game-ops/factory/racing"
	_ "github.com/neko233-com/game-ops/factory/rhythm"
	_ "github.com/neko233-com/game-ops/factory/roguelike"
	_ "github.com/neko233-com/game-ops/factory/rpg"
	_ "github.com/neko233-com/game-ops/factory/sandbox"
	_ "github.com/neko233-com/game-ops/factory/simulation"
	_ "github.com/neko233-com/game-ops/factory/slg"
	_ "github.com/neko233-com/game-ops/factory/sports"
	_ "github.com/neko233-com/game-ops/factory/strategy"
	_ "github.com/neko233-com/game-ops/factory/survival"
	_ "github.com/neko233-com/game-ops/factory/tower_defense"
	_ "github.com/neko233-com/game-ops/factory/tps"
	_ "github.com/neko233-com/game-ops/factory/visual_novel"
)
