package hook

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/wini"
)

// All available hook groups. Each hook group corresponds to some action in
// Wingo. When that action happens, every hook in the corresponding group
// is fired.
const (
	Startup   Type = "startup"
	Restarted Type = "restart"
	Managed   Type = "managed"
	Focused   Type = "focused"
	Unfocused Type = "unfocused"
)

var (
	// A global corresponding to the Gribble execution environment.
	gribbleEnv *gribble.Environment

	// A map from group constants to group values.
	groups = map[Type]group{
		Startup:   make(group, 0),
		Restarted: make(group, 0),
		Managed:   make(group, 0),
		Focused:   make(group, 0),
		Unfocused: make(group, 0),
	}
)

type Type string

type group []hook

type hook struct {
	// From the config file. Nice for error messages.
	name string

	// List of match conditions expressed as Gribble commands.
	satisfies []string

	// When true, each condition in 'satisfies' will be combined as a set
	// of conjunctions.
	// When false, each condition in 'satisfies' will be combined as a set
	// of disjunctions.
	conjunction bool

	// List of consequences expressed as Gribble commands. All consequences
	// are fired if '(and satisfies[0] satisfies[1] ... satisfies[n-1])' is
	// satisfied.
	consequences []string
}

// Initializes the hooks package with a Gribble execution environment and
// a file path to a wini formatted hooks configuration file. If the
// initialization fails, only a warning is logged since hooks are not
// essential for Wingo to run.
func Initialize(env *gribble.Environment, fpath string) {
	gribbleEnv = env

	cdata, err := wini.Parse(fpath)
	if err != nil {
		logger.Warning.Printf("Could not parse '%s': %s", fpath, err)
		return
	}
	for _, hookName := range cdata.Sections() {
		if err := readSection(cdata, hookName); err != nil {
			logger.Warning.Printf("Could not load hook '%s': %s", hookName, err)
		}
	}
}

// Fire will attempt to run every hook in the group specified, while replacing
// special strings in every command executed with those provided by args.
//
// Running a hook is a two step process. First, the match conditions are
// executed. If all match conditions return true, we proceed to execute all
// of the consequences. If any of the match conditions are false, we stop
// and condinue on to the next hook.
//
// Note that Fire immediately returns, as it executes in its own goroutine.
func Fire(hk Type, args Args) {
	go func() {
		if _, ok := groups[hk]; !ok {
			logger.Warning.Printf("Unknown hook group '%s'.", hk)
			return
		}
		for _, hook := range groups[hk] {
			// Run all of the match conditions. Depending upon the value
			// of hk.conjunction, we treat the conditions as either a set
			// of conjunctions or a set of disjunctions.
			andMatched := true
			orMatched := false
			for _, condCmd := range hook.satisfies {
				val, err := gribbleEnv.Run(args.apply(condCmd))
				if err != nil {
					logger.Warning.Printf("When executing the 'match' "+
						"conditions for your '%s' hook in the '%s' group, "+
						"the command '%s' returned an error: %s",
						hook.name, hk, condCmd, err)
					andMatched = false
					orMatched = false
					break
				}
				if gribbleBool(val) {
					logger.Lots.Printf("Condition '%s' matched "+
						"for the hook '%s' in the '%s' group.",
						condCmd, hook.name, hk)
					orMatched = true
					if !hook.conjunction {
						break
					}
				} else {
					logger.Lots.Printf("Condition '%s' failed to match "+
						"for the hook '%s' in the '%s' group.",
						condCmd, hook.name, hk)
					andMatched = false
					if hook.conjunction {
						break
					}
				}
			}
			if hook.conjunction && !andMatched {
				continue
			}
			if !hook.conjunction && !orMatched {
				continue
			}

			logger.Lots.Printf("The hook '%s' in the '%s' group has matched!",
				hook.name, hk)

			// We have a match! Let's proceed to the consequences...
			for _, consequentCmd := range hook.consequences {
				_, err := gribbleEnv.Run(args.apply(consequentCmd))
				if err != nil {
					logger.Warning.Printf("When executing the consequences "+
						"for your '%s' hook in the '%s' group, the command "+
						"'%s' returned an error: %s",
						hook.name, hk, consequentCmd, err)
					// consequent commands are independent, so we march on.
				}
			}
		}
	}()
}

// gribbleBool translates a value returned by a Gribble command to a boolean
// value. A command returns true if and only if it returns the integer 1.
// Any other value results in false.
func gribbleBool(val gribble.Any) bool {
	if v, ok := val.(int); ok && v == 1 {
		return true
	}
	return false
}

// readSection loads a particular section from the configuration file into
// the hook groups. One section may result in the same hook being added to
// multiple groups.
func readSection(cdata *wini.Data, section string) error {
	// First lets roll up the match conditions.
	match := cdata.GetKey(section, "match")
	if match == nil {
		return fmt.Errorf("Could not find 'match' in the '%s' hook.", section)
	}

	satisfies := make([]string, len(match.Strings()))
	copy(satisfies, match.Strings())

	// Check each satisfies command to make sure it's a valid Gribble command.
	if cmd, err := checkCommands(satisfies); err != nil {
		return fmt.Errorf("The match command '%s' in the '%s' hook could "+
			"not be parsed: %s", cmd, section, err)
	}

	// Now try to find whether it's a conjunction or not.
	conjunction := true
	conjunctionKey := cdata.GetKey(section, "conjunction")
	if conjunctionKey != nil {
		if vals, err := conjunctionKey.Bools(); err != nil {
			logger.Warning.Println(err)
		} else {
			conjunction = vals[0]
		}
	}

	// Now traverse all of the keys in the section. We'll skip "match" since
	// we've already grabbed the data. Any other key should correspond to
	// a hook group name.
	addedOne := false
	for _, key := range cdata.Keys(section) {
		groupName := Type(key.Name())
		if groupName == "match" || groupName == "conjunction" {
			continue
		}
		if _, ok := groups[groupName]; !ok {
			return fmt.Errorf("Unrecognized hook group '%s' in the '%s' hook.",
				groupName, section)
		}

		consequences := make([]string, len(key.Strings()))
		copy(consequences, key.Strings())

		// Check each consequent command to make sure it's valid.
		if cmd, err := checkCommands(consequences); err != nil {
			return fmt.Errorf("The '%s' command '%s' in the '%s' hook could "+
				"not be parsed: %s", groupName, cmd, section, err)
		}
		hook := hook{
			name:         section,
			satisfies:    satisfies,
			conjunction:  conjunction,
			consequences: consequences,
		}
		groups[groupName] = append(groups[groupName], hook)
		addedOne = true
	}
	if !addedOne {
		return fmt.Errorf(
			"No hook groups were detected in the '%s' hook.", section)
	}
	return nil
}

// checkCommands runs through a list of strings and tries to parse each as
// a Gribble command in 'gribbleEnv'. If an error occurs in any of them,
// the errant command and the error are returned.
func checkCommands(cmds []string) (string, error) {
	for _, cmd := range cmds {
		if err := gribbleEnv.Check(cmd); err != nil {
			return cmd, err
		}
	}
	return "", nil
}

// Args represents a value that specifies what special strings get replaced
// with in user defined hooks. So for instance, the "focused" hook is
// specifically defined on a particular client, so its hook is fired like so:
//
//	args := Args{
//		Client: "identifier of window being focused",
//	}
//	hook.Fire(hook.Focused, args)
type Args struct {
	Client string
}

// apply takes a command string and replaces special strings with values in
// Args that have non-zero length.
func (args Args) apply(cmd string) string {
	replace := make([]string, 0)
	if len(args.Client) > 0 {
		replace = append(replace, []string{"\":client:\"", args.Client}...)
	}

	if len(replace) == 0 {
		return cmd
	}
	return strings.NewReplacer(replace...).Replace(cmd)
}
