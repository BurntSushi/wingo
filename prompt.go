package main

type prompts struct {
	cycle *promptCycle
	slct  *promptSelect // temporary
}

func promptsInitialize() {
	PROMPTS = prompts{
		cycle: newPromptCycle(),
		slct:  newPromptSelect(),
	}
}

func (c *client) promptAdd() {
	c.promptCycleAdd()
	c.promptSelectAdd()
}

func (c *client) promptRemove() {
	c.promptCycleRemove()
	c.promptSelectRemove()
}

func (c *client) promptUpdateIcon() {
	c.promptCycleUpdateIcon()
}

func (c *client) promptUpdateName() {
	c.promptCycleUpdateName()
	c.promptSelectUpdateName()
}

func (wrk *workspace) promptAdd() {
	wrk.promptSelectAdd()
}

func (wrk *workspace) promptRemove() {
	wrk.promptSelectRemove()
}

func (wrk *workspace) promptUpdateName() {
	wrk.promptSelectUpdateName()
}
