// +build windows

package argon

import (
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/execdriver"
)

type TtyConsole struct {
	//	MasterPty *os.File
}

func NewTtyConsole(processConfig *execdriver.ProcessConfig, pipes *execdriver.Pipes) (*TtyConsole, error) {

	tty := &TtyConsole{}

	return tty, nil
}

func (t *TtyConsole) Resize(h, w int) error {
	log.Debugln("argon ttyconsole: resize not yet implemented ", h, w)
	//return term.SetWinsize(t.MasterPty.Fd(), &term.Winsize{Height: uint16(h), Width: uint16(w)})
	return nil
}

func (t *TtyConsole) Close() error {
	return nil
}
