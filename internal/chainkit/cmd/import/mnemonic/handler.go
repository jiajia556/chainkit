package mnemonic

import (
	"github.com/jiajia556/chainkit/models/chainkitmnemonics"
	"github.com/jiajia556/chainkit/pkg/utils"
)

func ImportMnemonic(words, password, remark string) {
	mnemonic := chainkitmnemonics.NewRecord()
	err := mnemonic.ImportMnemonic(words, password, remark)
	if err != nil {
		utils.OutputFatal("CreateAndGetNewMnemonic failed: ", err)
	}
}
