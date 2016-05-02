package terraform

import (
	"fmt"
	"log"
)

// EvalReadData is an EvalNode implementation that executes a data resource's
// ReadData method and populates its state.
type EvalReadData struct {
	Provider *ResourceProvider
	Output   **InstanceState
	Config   **ResourceConfig
	Info     *InstanceInfo
}

func (n *EvalReadData) Eval(ctx EvalContext) (interface{}, error) {
	// TODO: test
	provider := *n.Provider

	config := *n.Config
	if config == nil {
		// Should never happen
		panic(fmt.Errorf("EvalDataResourceInit for %s given nil ResourceConfig", n.Info.HumanId()))
	}

	// We can't initialize until our config has been completely interpolated.
	// If a data resource depends on a not-yet-created managed resource then
	// we'll exit here during Refresh and then visit again during Apply,
	// at which point the dependencies should all be ready.
	if config.ComputedKeys != nil && len(config.ComputedKeys) > 0 {
		log.Printf("[TRACE] %s: skipping read: config has computed attributes", n.Info.Id)
		return nil, nil
	}

	// Call pre-refresh hook
	err := ctx.Hook(func(h Hook) (HookAction, error) {
		// We don't have a state yet, so we'll just give the hook an
		// empty one to work with.
		return h.PreRefresh(n.Info, &InstanceState{})
	})
	if err != nil {
		return nil, err
	}

	// Refresh!
	state, err := provider.ReadData(n.Info, config)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", n.Info.Id, err.Error())
	}

	// Call post-refresh hook
	err = ctx.Hook(func(h Hook) (HookAction, error) {
		return h.PostRefresh(n.Info, state)
	})
	if err != nil {
		return nil, err
	}

	if n.Output != nil {
		*n.Output = state
	}

	return nil, nil
}
