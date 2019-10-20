package main

import (
	"echo-oidc-client/pkg/cli"
	"echo-oidc-client/pkg/globals"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
)

var (
	TempFolder = os.Getenv("TEMP") // windows
)

func printMainHelp() {
	fmt.Println("expected 'Login' subcommands")

}

func main() {
	if runtime.GOOS == "windows" {
		TempFolder = os.Getenv("TEMP")
	} else {
		TempFolder = os.Getenv("TMPDIR")
	}

	dir := path.Join(TempFolder, "_afx_cli")
	err, db := globals.OpenBadgerDb(dir)

	if err != nil {
		log.Fatalf("Could not open DB!: %s", err.Error())
		panic("Could not open DB!")
	}
	defer db.Close()

	fmt.Println(dir)
	loginCmd := flag.NewFlagSet("Login", flag.ExitOnError)
	loginState := loginCmd.Bool("state", false, "fetches the current state of the login")
	loginSignOut := loginCmd.Bool("sign-out", false, "deletes the local account data.  i.e. removes the access_token")
	loginHelp := loginCmd.Bool("h", false, "help")
	if len(os.Args) < 2 {
		printMainHelp()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "Login":
		loginCmd.Parse(os.Args[2:])

		if *loginHelp {
			fmt.Println("subcommand 'Login'")
			loginCmd.PrintDefaults()
		} else {
			if *loginState {
				cli.LoginState()
			}
			if *loginSignOut {
				cli.LoginSignOut()
			}

			if !*loginState && !*loginSignOut {
				cli.LoginSignOut()
				cli.Login()
			}

		}

	default:
		printMainHelp()
		os.Exit(1)
	}

}
