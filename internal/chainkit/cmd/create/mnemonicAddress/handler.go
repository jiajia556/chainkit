package mnemonicAddress

import (
	"fmt"

	"github.com/jiajia556/chainkit/models/chainkitmnemonicaddresses"
	"github.com/jiajia556/chainkit/pkg/utils"
)

func createMnemonicAddress(mneId uint64, index uint32, password, remark string) {
	addr := chainkitmnemonicaddresses.NewRecord().GetByMnemonicIdAndIndex(mneId, index)
	if addr.Exists() {
		utils.OutputFatal("Address already exists for the given mnemonic ID and index")
	}
	mneAddr := chainkitmnemonicaddresses.NewRecord()
	err := mneAddr.CreateByMnemonicAndIndex(mneId, index, password, remark)
	if err != nil {
		utils.OutputFatal("Failed to create mnemonic address", err)
	}
	fmt.Printf("id: %d, address: %s", mneAddr.Model.Id, mneAddr.Model.Address)
}
