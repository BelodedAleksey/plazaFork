package functions

import (
	"github.com/go-vgo/robotgo"
	"github.com/sirupsen/logrus"
)

func Key() {
	ok := robotgo.AddEvent("q")
	if ok {
		logrus.Info("Q PRESSED!")
	}
}
