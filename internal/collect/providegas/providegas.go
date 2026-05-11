package providegas

import (
	"context"

	"github.com/jiajia556/chainkit/models/chainkitchains"
	"github.com/jiajia556/chainkit/models/chainkitcollectgasfeetasks"
	"github.com/jiajia556/tool-box/log"
)

func Start(ctx context.Context) {
	chains := chainkitchains.NewList()
	err := chains.FindAll()
	if err != nil {
		log.Error("failed to find chains", "error", err)
		return
	}

	chains.Foreach(func(key int, chain *chainkitchains.Record) bool {

		return true
	})
}

func handleChain(chain *chainkitchains.Record) {
	waitingList := chainkitcollectgasfeetasks.NewList().GetWaitingList(chain.Model.Id)
	if waitingList.IsEmpty() {
		return
	}

	waitingList.Foreach(func(key int, gasTask *chainkitcollectgasfeetasks.Record) bool {

		return true
	})
}
