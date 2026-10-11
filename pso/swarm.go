package pso

import (
	"errors"
	"fmt"

	"context"

	"github.com/mfateev/sdk-go-poc/workflow"
)

type ParticleResult struct {
	Position
	Step int
}

type Swarm struct {
	Settings  *SwarmSettings
	Gbest     *Position
	Particles []*Particle
}

func NewSwarm(ctx context.Context, settings *SwarmSettings) (*Swarm, error) {
	var swarm Swarm
	// store settings
	swarm.Settings = settings
	// initialize gbest
	swarm.Gbest = NewPosition(swarm.Settings.function.dim)
	swarm.Gbest.Fitness = 1e20

	// Collect a whole batch before publishing any particle state.
	swarm.Particles = make([]*Particle, settings.Size)
	particles, err := swarm.runParticleActivities(ctx, InitParticleActivityName, false)
	if err != nil {
		return nil, err
	}
	swarm.Particles = particles

	swarm.updateBest()

	return &swarm, nil
}

func (swarm *Swarm) updateBest() {
	for i := 0; i < swarm.Settings.Size; i++ {
		if swarm.Particles[i].Pbest.IsBetterThan(swarm.Gbest) {
			swarm.Gbest = swarm.Particles[i].Pbest.Copy()
		}
	}
}

func (swarm *Swarm) Run(ctx context.Context, step int) (ParticleResult, error) {
	logger := workflow.GetLogger(ctx)

	// Setup query handler for query type "iteration"
	var iterationMessage string
	err := workflow.SetQueryHandler(ctx, "iteration", func() (string, error) {
		return iterationMessage, nil
	})
	if err != nil {
		logger.Info("SetQueryHandler failed: " + err.Error())
		return ParticleResult{}, err
	}

	for step <= swarm.Settings.Steps {
		logger.Info("Iteration", "step", step)
		particles, err := swarm.runParticleActivities(ctx, UpdateParticleActivityName, true)
		if err != nil {
			return ParticleResult{Position: *swarm.Gbest, Step: step}, err
		}
		swarm.Particles = particles

		logger.Debug("Iteration Update Swarm Best", "step", step)

		swarm.updateBest()

		// Check if the goal has reached then stop early
		if swarm.Gbest.Fitness < swarm.Settings.function.Goal {
			logger.Debug("Iteration New Swarm Best", "step", step)
			return ParticleResult{
				Position: *swarm.Gbest,
				Step:     step,
			}, nil
		}

		iterationMessage = fmt.Sprintf("Step %d :: min err=%.5e\n", step, swarm.Gbest.Fitness)
		if step%swarm.Settings.PrintEvery == 0 {
			logger.Info(iterationMessage)
		}

		// Finished all iterations
		if step == swarm.Settings.Steps {
			break
		}

		// Not finished yet, just continue as new to reduce history size
		if step%swarm.Settings.ContinueAsNewEvery == 0 {
			return ParticleResult{
				Position: *swarm.Gbest,
				Step:     step,
			}, errors.New(ContinueAsNewStr)
		}

		step++
	}

	return ParticleResult{
		Position: *swarm.Gbest,
		Step:     step,
	}, nil
}

// Every activity in a batch sees the same swarm snapshot. Only the collecting
// goroutine updates the particle slice, after all results have arrived.
func (swarm *Swarm) runParticleActivities(ctx context.Context, activityName string, indexed bool) ([]*Particle, error) {
	type result struct {
		index    int
		particle Particle
		err      error
	}
	results := make(chan result, swarm.Settings.Size)
	for index := range swarm.Settings.Size {
		go func() {
			value := result{index: index}
			if indexed {
				value.err = workflow.ExecuteActivity(ctx, activityName, *swarm, index).Get(ctx, &value.particle)
			} else {
				value.err = workflow.ExecuteActivity(ctx, activityName, *swarm).Get(ctx, &value.particle)
			}
			results <- value
		}()
	}
	particles := make([]*Particle, swarm.Settings.Size)
	for range swarm.Settings.Size {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case value := <-results:
			if value.err != nil {
				return nil, value.err
			}
			particles[value.index] = &value.particle
		}
	}
	return particles, nil
}
