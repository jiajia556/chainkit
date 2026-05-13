package depaddr

import (
	"github.com/jiajia556/chainkit/models/chainkituserdepositaddress"
	"github.com/jiajia556/chainkit/pkg/utils"
)

func createDepositAddress(count int, password, remark string) {
	createdCount, err := chainkituserdepositaddress.NewRecord().BatchCreate(count, password, remark)
	if err != nil {
		utils.OutputFatal(
			"Create deposit address error: ", err,
			"created count: ", createdCount,
		)
	}
}
