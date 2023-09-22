package events

import (
	"fmt"

	"github.com/3atlab/netroub/pkg/model"
)

func ExecuteEvent(index int) error {
	switch model.Scenar.Event[index].Type {
	case "pumba":
		err := ExecPumbaCommand(index)
		if err != nil {
			return err
		}
	case "config":
		err := ExecConfigCommand(index)
		if err != nil {
			return err
		}
	default:
		fmt.Println("Invalid event type")
	}
	return nil
}
