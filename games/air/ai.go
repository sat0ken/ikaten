package air

// The enemy pilots' tasks.
const (
	taskPatrol = iota // circle over the fleet until the player comes near
	taskEscort        // fly with the bombers, and turn on the player when close
	taskAttack        // chase the player and shoot
	taskBreak         // turn hard away from the player behind
	taskRun           // a bomber flying straight at the fleet
	taskLeave         // fly away and be gone
)

const (
	gunRange    = 550 // meters at which the enemy fires
	spotRange   = 4500
	escortRange = 2500
	pullUpAlt   = 400
)

// think flies enemy plane i for one tick.
func (g *Game) think(i int) {
	q := &g.planes[i]
	p := g.player()
	if q.pos.z <= 0 {
		q.alive = false
		g.addPuff(q.pos.flat(), puffSplash, 20, 90)
		return
	}
	if q.falling > 0 {
		q.falling++
		if q.falling%5 == 0 {
			g.addPuff(q.pos, puffSmoke, 5, 80)
		}
		q.fly(0.7, -0.3)
		return
	}
	if q.gunCool > 0 {
		q.gunCool--
	}
	playerUp := p.alive && p.falling == 0
	toPlayer := p.pos.sub(q.pos)
	dist := toPlayer.length()
	fleet := g.ships[0].pos
	switch q.task {
	case taskPatrol:
		if playerUp && p.pos.flat().sub(fleet).length() < spotRange {
			q.task = taskAttack
		}
	case taskEscort:
		if playerUp && dist < escortRange {
			q.task = taskAttack
		}
	case taskAttack:
		if !playerUp {
			q.task = taskLeave
		}
		// The player on its tail makes it break.
		behind := q.pos.sub(p.pos)
		if dist < 700 && behind.norm().dot(p.f) > 0.97 && toPlayer.norm().dot(q.f) < 0 && g.random(90) == 0 {
			q.task, q.until = taskBreak, g.clock+120+g.random(90)
			q.breakDir = float32(g.random(2)*2 - 1)
		}
	case taskBreak:
		if g.clock > q.until {
			q.task = taskAttack
		}
		q.lastAim = vec{}
	}

	var roll, pull float32
	switch q.task {
	case taskPatrol:
		// A wide circle around the fleet.
		c := fleet.add(heading(float32(g.clock)*0.002 + float32(i)).mul(1800))
		c.z = 2000
		roll, pull = q.steer(c)
	case taskEscort:
		lead := g.leadBomber()
		if lead >= 0 {
			b := &g.planes[lead]
			roll, pull = q.steer(b.pos.add(b.f.mul(300)).add(vec{0, 0, 300}))
		} else {
			q.task = taskLeave
		}
	case taskAttack:
		// Lead the player by the bullets' time of flight, with an error that
		// changes now and then.
		if g.clock%120 == i*7%120 {
			q.aimErr = vec{g.spread(), g.spread(), g.spread()}.mul(6)
		}
		tof := dist / bulletSpeed
		aim := p.pos.add(p.f.mul(p.v * tof)).add(q.aimErr)
		// Steer where the aim point is going, as it swings across the nose.
		if q.lastAim != (vec{}) {
			roll, pull = q.steer(aim.add(aim.sub(q.lastAim).mul(0.4 * tps)))
		} else {
			roll, pull = q.steer(aim)
		}
		q.lastAim = aim
		if dist < 120 && toPlayer.dot(q.f) > 0 {
			// Too close: pass by.
			q.task, q.until, q.breakDir = taskBreak, g.clock+60, 1
		}
		// The guns converge a little: they fire along the aim within three
		// degrees of the nose.
		if dir := aim.sub(q.pos).norm(); q.gunCool == 0 && dist < gunRange && dir.dot(q.f) > 0.9986 {
			g.shoot(q, dir, bulletSpeed, 1, 0.025, true)
			q.gunCool = 8
			// Short bursts, then a pause.
			if q.burst++; q.burst%6 == 0 {
				q.gunCool = 50
			}
		}
	case taskBreak:
		roll, pull = q.breakDir, 1
		if q.bank()*q.breakDir > 1.3 {
			roll = 0
		}
	case taskRun:
		goal := q.goal
		goal.z = q.pos.z
		roll, pull = q.steer(goal)
		if q.pos.flat().sub(q.goal).length() < 500 {
			g.raidHit(q)
		}
		g.gunner(q, false)
	case taskLeave:
		away := q.pos.add(q.pos.sub(fleet).flat().norm().mul(3000))
		away.z = 3000
		roll, pull = q.steer(away)
		if q.pos.flat().sub(fleet).length() > leaveRange {
			q.alive = false
		}
		if q.kind == bomber {
			g.gunner(q, false)
		}
	}
	// Keep clear of the sea, with room for the dive to round out.
	if q.pos.z < pullUpAlt+max(0, -q.f.z)*q.v*4 && q.f.z < 0.2 {
		roll, pull = q.level(), 1
	}
	q.fly(roll, pull)
}

// leadBomber returns the first bomber still flying, or -1.
func (g *Game) leadBomber() int {
	for i := 1; i < maxPlanes; i++ {
		if q := &g.planes[i]; q.alive && q.falling == 0 && q.kind == bomber {
			return i
		}
	}
	return -1
}

// raidHit is a bomber over the carrier: it drops its bombs and turns away.
func (g *Game) raidHit(q *plane) {
	q.task = taskLeave
	c := &g.ships[0]
	if g.random(100) < 45 && c.hp > 0 {
		c.hp--
		c.fire = 1200
		g.addPuff(c.pos.add(vec{0, 0, 10}), puffFire, 40, 90)
		g.say(msgCarrierHit)
	} else {
		g.addPuff(c.pos.add(heading(float32(g.random(628))/100).mul(120)), puffSplash, 30, 90)
	}
}

// gunner fires a plane's rear gun at the target behind it: the enemy
// bomber's at the player, or the player's at the enemy fighters.
func (g *Game) gunner(q *plane, player bool) {
	if player {
		if g.gunnerCool > 0 {
			g.gunnerCool--
			return
		}
		for i := 1; i < maxPlanes; i++ {
			e := &g.planes[i]
			if !e.alive || e.falling > 0 || e.kind != fighter {
				continue
			}
			d := e.pos.sub(q.pos)
			if dl := d.length(); dl < 450 && d.dot(q.f) < 0 {
				tof := dl / bulletSpeed
				aim := e.pos.add(e.f.mul(e.v * tof)).sub(q.pos).norm()
				g.shoot(q, aim, bulletSpeed, 2, 0.02, false)
				g.gunnerCool = 9
				return
			}
		}
		return
	}
	p := g.player()
	if !p.alive || p.falling > 0 || q.gunCool > 0 {
		return
	}
	d := p.pos.sub(q.pos)
	if dl := d.length(); dl < 400 {
		tof := dl / bulletSpeed
		aim := p.pos.add(p.f.mul(p.v * tof)).sub(q.pos).norm()
		g.shoot(q, aim, bulletSpeed, 1, 0.05, true)
		q.gunCool = 15
	}
}
