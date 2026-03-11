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
	case "auth":
		return a.runSubmitAuth(args[1:])
	case "week":
		return a.runSubmitWeek(args[1:])
	default:
		fmt.Fprint(a.Stdout, submitHelp)
		return &exitError{Code: 2, Message: fmt.Sprintf("unknown submit subcommand %q", args[0])}
	}
}

func (a *App) runSubmitAuth(args []string) error {
	if len(args) == 0 {
		fmt.Fprint(a.Stdout, submitAuthHelp)
		return &exitError{Code: 2, Message: "missing submit auth subcommand"}
	}

	switch args[0] {
	case "help":
		fmt.Fprint(a.Stdout, submitAuthHelp)
		return nil
	case "login":
		return a.runSubmitAuthLogin(args[1:])
	case "status":
		return a.runSubmitAuthStatus(args[1:])
	case "logout":
		return a.runSubmitAuthLogout(args[1:])
	default:
		fmt.Fprint(a.Stdout, submitAuthHelp)
		return &exitError{Code: 2, Message: fmt.Sprintf("unknown submit auth subcommand %q", args[0])}
	}
}
