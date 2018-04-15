package plugin

import (
	"github.com/ipfs/go-ipfs/core/coreapi/interface"
)

type APIConsumer interface {
	ConsumeAPI(api iface.CoreAPI)
}
