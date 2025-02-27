package test

import (
	"context"
	"fmt"

	"github.com/JCorpse96/core/action"
	"github.com/JCorpse96/core/data/expression"
	"github.com/JCorpse96/core/data/mapper"
	"github.com/JCorpse96/core/data/metadata"
	"github.com/JCorpse96/core/data/resolve"
	"github.com/JCorpse96/core/engine/runner"
	"github.com/JCorpse96/core/support/log"
	"github.com/JCorpse96/core/trigger"
)

func InitTrigger(factory trigger.Factory, tConfig *trigger.Config, actions map[string]action.Action) (trigger.Trigger, error) {

	r := runner.NewDirect()

	if factory == nil {
		return nil, fmt.Errorf("trigger factory not provided")
	}

	trg, err := factory.New(tConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot create trigger '%s': %v", tConfig.Id, err)
	}
	if trg == nil {
		return nil, fmt.Errorf("cannot create trigger '%s'", tConfig.Id)
	}

	err = tConfig.FixUp(trigger.NewMetadata(), resolve.GetBasicResolver())
	if err != nil {
		return nil, fmt.Errorf("cannot create trigger '%s': %v", tConfig.Id, err)
	}

	mf := mapper.NewFactory(resolve.GetBasicResolver())
	ef := expression.NewFactory(resolve.GetBasicResolver())

	initCtx := initContext{handlers: make([]trigger.Handler, 0, len(tConfig.Handlers)), logger: logger}
	var acts []action.Action

	//create handlers for that trigger and init
	for _, hConfig := range tConfig.Handlers {

		id := hConfig.Actions[0].Id
		act := actions[id]

		acts = append(acts, act)

		handler, _ := trigger.NewHandler(hConfig, acts, mf, ef, r, logger)

		initCtx.handlers = append(initCtx.handlers, handler)

	}

	err = trg.Initialize(initCtx)
	if err != nil {
		return nil, err
	}

	return trg, nil
}

//////////////////////////
// Simple Init Context

type initContext struct {
	handlers []trigger.Handler
	logger   log.Logger
}

func (ctx initContext) GetHandlers() []trigger.Handler {
	return ctx.handlers
}
func (ctx initContext) Logger() log.Logger {
	return ctx.logger
}

//////////////////////////
// Dummy Test Action

func NewDummyAction(f func()) action.Action {
	return &testAction{f: f}
}

type testAction struct {
	f func()
}

func (a *testAction) IOMetadata() *metadata.IOMetadata {
	return nil
}

// Metadata get the Action's metadata
func (a *testAction) Metadata() *action.Metadata {
	return nil
}

// Run implementation of action.SyncAction.Run
func (a *testAction) Run(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	a.f()
	return nil, nil
}
