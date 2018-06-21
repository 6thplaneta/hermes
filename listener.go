package hermes

import (
	"os"
)

//
type Listener interface {
	OnTerminate(os.Signal)
}
