package shooter

import "testing"

func playing() *Game {
	g := NewGame()
	g.setMode(modePlay)
	g.waveTick = -1 << 20 // no waves unless a test starts them
	return g
}

func TestShotDestroysScout(t *testing.T) {
	g := playing()
	g.spawn(scout, 60)
	g.enemies[0].y, g.enemies[0].vy = 40*fp, 0
	g.px, g.py = 60*fp, 100*fp
	g.addShot(0, 0)
	for range 30 {
		g.updateMovers()
		g.collide()
	}
	if g.enemies[0].alive || g.score != 100 || g.kills != 1 {
		t.Fatalf("alive=%v score=%d kills=%d", g.enemies[0].alive, g.score, g.kills)
	}
}

func TestBombClearsBulletsAndDamages(t *testing.T) {
	g := playing()
	g.spawn(weaver, 60)
	g.enemies[0].y = 30 * fp
	g.fan(60*fp, 30*fp, down, 5, 3, 20)
	g.bomb()
	if g.enemies[0].alive || g.bombs != startBombs-1 || g.invincible == 0 {
		t.Fatalf("alive=%v bombs=%d invincible=%d", g.enemies[0].alive, g.bombs, g.invincible)
	}
	for _, b := range g.bullets {
		if b.alive {
			t.Fatal("bullet survived the bomb")
		}
	}
}

func TestDeathsEndTheGame(t *testing.T) {
	g := playing()
	for death := 1; death <= startLives; death++ {
		for g.invincible > 0 || g.respawn > 0 {
			g.updatePlayer()
		}
		g.fan(g.px, g.py, down, 1, 0, 0) // a bullet resting on the cockpit
		g.collide()
		if g.lives != startLives-death || g.respawn == 0 {
			t.Fatalf("death %d: lives=%d respawn=%d", death, g.lives, g.respawn)
		}
		clear(g.bullets[:])
	}
	for range 90 {
		g.updatePlayer()
	}
	if g.mode != modeOver {
		t.Fatalf("mode=%d after the last life", g.mode)
	}
}

func TestWavesLeadToBossAndNextStage(t *testing.T) {
	g := playing()
	g.waveTick = 0
	var bossAt int
	for tick := 1; tick < 60*120 && bossAt == 0; tick++ {
		g.invincible = 2 // survive while the waves pass
		if err := g.Update(); err != nil {
			t.Fatal(err)
		}
		for i := range g.enemies {
			if e := &g.enemies[i]; e.alive && e.kind == boss {
				bossAt = tick
			}
		}
	}
	if bossAt == 0 {
		t.Fatal("no boss after the waves")
	}
	for i := range g.enemies {
		if e := &g.enemies[i]; e.alive && e.kind == boss {
			g.damage(e, e.hp)
		}
	}
	if g.level != 2 || g.wave != 0 || g.bossSpawned {
		t.Fatalf("level=%d wave=%d bossSpawned=%v", g.level, g.wave, g.bossSpawned)
	}
	t.Logf("boss after %d ticks (%.0f s)", bossAt, float64(bossAt)/60)
}

func TestAimPointsAtPlayer(t *testing.T) {
	g := playing()
	g.px, g.py = 60*fp, 100*fp
	if got := g.aim(60*fp, 20*fp); got != down {
		t.Fatalf("straight down: %d", got)
	}
	if got := g.aim(100*fp, 100*fp); got != 32 {
		t.Fatalf("to the left: %d", got)
	}
}
