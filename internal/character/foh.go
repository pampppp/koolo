package character

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/helper"
	"github.com/hectorgimenez/koolo/internal/pather"
	"sort"
	"time"
)

const (
	fohMaxAttacksLoop = 20
	fohMinDistance    = 5
	fohMaxDistance    = 20
)

type Foh struct {
	BaseCharacter
}

func (s Foh) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
	opts ...step.AttackOption,
) action.Action {
	completedAttackLoops := 0
	previousUnitID := 0

	return action.NewStepChain(func(d game.Data) []step.Step {
		id, found := monsterSelector(d)
		if !found {
			return []step.Step{}
		}
		if previousUnitID != int(id) {
			completedAttackLoops = 0
		}

		if !s.preBattleChecks(d, id, skipOnImmunities) {
			return []step.Step{}
		}

		if completedAttackLoops >= fohMaxAttacksLoop {
			return []step.Step{}
		}

		steps := []step.Step{
			step.PrimaryAttack(
				id,
				3,
				step.Distance(fohMinDistance, fohMaxDistance),
				step.EnsureAura(skill.Conviction),
			),
		}

		completedAttackLoops++
		previousUnitID = int(id)

		return steps
	}, action.RepeatUntilNoSteps())
}

func (s Foh) BuffSkills(_ game.Data) []skill.ID {
	return []skill.ID{skill.HolyShield}
}

func (s Foh) KillCountess() action.Action {
	return s.killMonster(npc.DarkStalker, data.MonsterTypeSuperUnique)
}

func (s Foh) KillAndariel() action.Action {
	return s.killMonster(npc.Andariel, data.MonsterTypeNone)
}

func (s Foh) KillSummoner() action.Action {
	return s.killMonster(npc.Summoner, data.MonsterTypeNone)
}

func (s Foh) KillDuriel() action.Action {
	return s.killMonster(npc.Duriel, data.MonsterTypeNone)
}

func (s Foh) KillPindle(_ []stat.Resist) action.Action {
	return s.killMonster(npc.DefiledWarrior, data.MonsterTypeSuperUnique)
}

func (s Foh) KillMephisto() action.Action {
	return s.killMonster(npc.Mephisto, data.MonsterTypeNone)
}

func (s Foh) KillNihlathak() action.Action {
	return s.killMonster(npc.Nihlathak, data.MonsterTypeSuperUnique)
}

func (s Foh) KillDiablo() action.Action {
	timeout := time.Second * 20
	startTime := time.Time{}
	diabloFound := false
	return action.NewChain(func(d game.Data) []action.Action {
		if startTime.IsZero() {
			startTime = time.Now()
		}

		if time.Since(startTime) > timeout && !diabloFound {
			s.logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := d.Monsters.FindOne(npc.Diablo, data.MonsterTypeNone)
		if !found || diablo.Stats[stat.Life] <= 0 {
			// Already dead
			if diabloFound {
				return nil
			}

			// Keep waiting...
			return []action.Action{action.NewStepChain(func(d game.Data) []step.Step {
				return []step.Step{step.Wait(time.Millisecond * 100)}
			})}
		}

		diabloFound = true
		s.logger.Info("Diablo detected, attacking")

		return []action.Action{
			s.killMonster(npc.Diablo, data.MonsterTypeNone),
			s.killMonster(npc.Diablo, data.MonsterTypeNone),
			s.killMonster(npc.Diablo, data.MonsterTypeNone),
		}
	}, action.RepeatUntilNoSteps())
}

func (s Foh) KillIzual() action.Action {
	return s.killMonster(npc.Izual, data.MonsterTypeNone)
}

func (s Foh) KillCouncil() action.Action {
	return action.NewStepChain(func(d game.Data) (steps []step.Step) {
		// Exclude monsters that are not council members
		var councilMembers []data.Monster
		for _, m := range d.Monsters {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				councilMembers = append(councilMembers, m)
			}
		}

		// Order council members by distance
		sort.Slice(councilMembers, func(i, j int) bool {
			distanceI := pather.DistanceFromMe(d, councilMembers[i].Position)
			distanceJ := pather.DistanceFromMe(d, councilMembers[j].Position)

			return distanceI < distanceJ
		})

		for _, m := range councilMembers {
			for range fohMaxAttacksLoop {
				steps = append(steps,
					step.PrimaryAttack(
						m.UnitID,
						3,
						step.Distance(fohMinDistance, fohMaxDistance),
						step.EnsureAura(skill.Conviction),
					),
				)
			}
		}
		return
	}, action.CanBeSkipped())
}

func (s Foh) KillBaal() action.Action {
	return s.killMonster(npc.BaalCrab, data.MonsterTypeNone)
}

func (s Foh) killMonster(npc npc.ID, t data.MonsterType) action.Action {
	return action.NewStepChain(func(d game.Data) (steps []step.Step) {
		m, found := d.Monsters.FindOne(npc, t)
		if !found {
			return nil
		}

		helper.Sleep(100)
		for range fohMaxAttacksLoop {
			steps = append(steps,
				step.PrimaryAttack(
					m.UnitID,
					3,
					step.Distance(fohMinDistance, fohMaxDistance),
					step.EnsureAura(skill.Conviction),
				),
			)
		}

		return
	}, action.CanBeSkipped())
}
