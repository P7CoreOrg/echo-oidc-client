/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"echo-oidc-client/afx/cmd"
	"os"
	"path"
	"runtime"

	"echo-oidc-client/pkg/globals"
	"log"
)

var (
	TempFolder = os.Getenv("TEMP") // windows
)

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

	cmd.Execute()
}
