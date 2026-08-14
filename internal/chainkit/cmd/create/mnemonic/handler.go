package mnemonic

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jiajia556/chainkit/models/chainkitmnemonics"
	"github.com/jiajia556/chainkit/pkg/utils"
)

func createMnemonic(password, out, remark string) {
	mnemonic := chainkitmnemonics.NewRecord()
	words, err := mnemonic.CreateAndGetNewMnemonic(password, remark)
	if err != nil {
		utils.OutputFatal("CreateAndGetNewMnemonic failed: ", err)
	}
	if out == "" {
		return
	}
	outMsg := fmt.Sprintf("id: %d,\nwords: %s", mnemonic.Model.Id, words)
	if out == "std" {
		fmt.Println(outMsg)
	} else {
		if filepath.Dir(out) != "." || filepath.Ext(out) != "" {
			if err := os.WriteFile(out, []byte(outMsg), 0o600); err != nil {
				utils.OutputFatal("write mnemonic output failed: ", err)
			}
		}
	}
	fmt.Println("Please store your mnemonic securely. Anyone with these words can access your funds.")
}
