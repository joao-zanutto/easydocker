package shared

import (
	"charm.land/bubbles/v2/key"
)

func helpKeyLabel(label string) string {
	return " " + label + " "
}

func UpBinding(help string) key.Binding {
	return key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", help),
	)
}

func DownBinding(help string) key.Binding {
	return key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", help),
	)
}

func LeftBinding(help string) key.Binding {
	return key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", help),
	)
}

func RightBinding(help string) key.Binding {
	return key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", help),
	)
}

func PageUpBinding(help string) key.Binding {
	return key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("pgup", help),
	)
}

func PageDownBinding(help string) key.Binding {
	return key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("pgdn", help),
	)
}

func HomeBinding(help string) key.Binding {
	return key.NewBinding(
		key.WithKeys("home"),
		key.WithHelp("home", help),
	)
}

func EndBinding(help string) key.Binding {
	return key.NewBinding(
		key.WithKeys("end"),
		key.WithHelp("end", help),
	)
}

func ActionBinding(keyName, help string) key.Binding {
	return key.NewBinding(
		key.WithKeys(keyName),
		key.WithHelp(helpKeyLabel(keyName), help),
	)
}

func EnterBinding(help string) key.Binding {
	return ActionBinding("enter", help)
}

func EscBinding(help string) key.Binding {
	return ActionBinding("esc", help)
}

func SlashBinding(help string) key.Binding {
	return ActionBinding("/", help)
}

func CanOpenShell(state string) bool {
	// Only running containers support shell execution
	return state == "running"
}
