package commands

import "fmt"

func (a *App) runSubmit(args []string) error {
	if len(args) == 0 {
		fmt.Fprint(a.Stdout, submitHelp)
		return &exitError{Code: 2, Message: "missing submit subcommand"}
	}

	switch args[0] {
	case "help":
		fmt.Fprint(a.Stdout, submitHelp)
		return nil
	case "week":
		return a.runSubmitWeek(args[1:])
	default:
		fmt.Fprint(a.Stdout, submitHelp)
		return &exitError{Code: 2, Message: fmt.Sprintf("unknown submit subcommand %q", args[0])}
	}
}
