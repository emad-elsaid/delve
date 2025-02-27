package terminal

import (
	"fmt"
	"os"
	"reflect"
	"text/tabwriter"

	"github.com/emad-elsaid/delve/pkg/config"
)

func configureCmd(t *Term, ctx callContext, args string) error {
	switch args {
	case "-list":
		return configureList(t)
	case "-save":
		return config.SaveConfig(t.conf)
	case "":
		return fmt.Errorf("wrong number of arguments to \"config\"")
	default:
		err := configureSet(t, args)
		if err != nil {
			return err
		}
		if t.client != nil { // only happens in tests
			lcfg := t.loadConfig()
			t.client.SetReturnValuesLoadConfig(&lcfg)
		}
		return nil
	}
}

func configureList(t *Term) error {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, ' ', 0)
	config.ConfigureList(w, t.conf, "yaml")
	return w.Flush()
}

func configureSet(t *Term, args string) error {
	v := config.Split2PartsBySpace(args)

	cfgname := v[0]
	var rest string
	if len(v) == 2 {
		rest = v[1]
	}

	if cfgname == "alias" {
		return configureSetAlias(t, rest)
	}

	field := config.ConfigureFindFieldByName(t.conf, cfgname, "yaml")
	if !field.CanAddr() {
		return fmt.Errorf("%q is not a configuration parameter", cfgname)
	}

	if field.Kind() == reflect.Slice && field.Type().Elem().Name() == "SubstitutePathRule" {
		return configureSetSubstitutePath(t, rest)
	}

	return config.ConfigureSetSimple(rest, cfgname, field)
}

func configureSetSubstitutePath(t *Term, rest string) error {
	argv := config.SplitQuotedFields(rest, '"')
	switch len(argv) {
	case 1: // delete substitute-path rule
		for i := range t.conf.SubstitutePath {
			if t.conf.SubstitutePath[i].From == argv[0] {
				copy(t.conf.SubstitutePath[i:], t.conf.SubstitutePath[i+1:])
				t.conf.SubstitutePath = t.conf.SubstitutePath[:len(t.conf.SubstitutePath)-1]
				return nil
			}
		}
		return fmt.Errorf("could not find rule for %q", argv[0])
	case 2: // add substitute-path rule
		for i := range t.conf.SubstitutePath {
			if t.conf.SubstitutePath[i].From == argv[0] {
				t.conf.SubstitutePath[i].To = argv[1]
				return nil
			}
		}
		t.conf.SubstitutePath = append(t.conf.SubstitutePath, config.SubstitutePathRule{From: argv[0], To: argv[1]})
	default:
		return fmt.Errorf("too many arguments to \"config substitute-path\"")
	}
	return nil
}

func configureSetAlias(t *Term, rest string) error {
	argv := config.SplitQuotedFields(rest, '"')
	switch len(argv) {
	case 1: // delete alias rule
		for k := range t.conf.Aliases {
			v := t.conf.Aliases[k]
			for i := range v {
				if v[i] == argv[0] {
					copy(v[i:], v[i+1:])
					t.conf.Aliases[k] = v[:len(v)-1]
				}
			}
		}
	case 2: // add alias rule
		alias, cmd := argv[1], argv[0]
		if t.conf.Aliases == nil {
			t.conf.Aliases = make(map[string][]string)
		}
		t.conf.Aliases[cmd] = append(t.conf.Aliases[cmd], alias)
	}
	t.cmds.Merge(t.conf.Aliases)
	return nil
}
