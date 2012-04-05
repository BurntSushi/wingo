package main

type prompts struct {
    cycle *promptCycle
}

func promptsInitialize() {
    PROMPTS = prompts{
        cycle: newPromptCycle(),
    }
}

func (c *client) promptAdd() {
    c.promptCycleAdd()
}

func (c *client) promptDestroy() {
    c.promptCycleDestroy()
}

func (c *client) promptUpdateIcon() {
    c.promptCycleUpdateIcon()
}

func (c *client) promptUpdateName() {
    c.promptCycleUpdateName()
}

