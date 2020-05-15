package sipengine

import (
	"context"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type Dialog struct {
	CallID string
	From string
	To string
	Start time.Time
	End time.Time
	Active bool
	//Metadata
	Labels map[string][]string
	Mux sync.Mutex
}


type DialogStore interface {
	Exists(callID string) error
	WriteDialog(d Dialog)
	UpdateDialog(d Dialog)
	DeleteDialog(callID string)
	ExistsByLabel(label string) error
	Details(callID string) Dialog
	Dialogs(callID string) []Dialog
	DialogsByLabel(label string) []Dialog
	Cancel(callID string)
	CancelByLabel(label string)
	//Start and prep for processing. This is also how
	//we would propagate certain error types up. IE
	//Shutdown errors up the stack on too many failures
	Boot() error
}

type InMemoryDialogStore struct {
	//Used to inherit cancellation et all
	//From upstream
	Ctx context.Context
	Dialogs []Dialog
	Channels struct {
		Write  chan Dialog
		ReadRequest chan string
		ReadResponse chan string
		Remove chan string
		Cancel chan string
		Error  chan error
	}
}


func (i *InMemoryDialogStore) Exists(callID string) error {

	for _, v := range i.Dialogs {
		if callID == v.CallID {
			return nil
		}
	}

	return errors.New("no matching dialog found")
}

func (i *InMemoryDialogStore) WriteDialog(d Dialog) {
	i.Channels.Write <- d
}

func (i *InMemoryDialogStore) ExistsByLabel(callID string) error {
	for _ , v := range i.Dialogs {
		for _ , l := range v.Labels {
			if len(l) > 0 {
				return nil
			}
		}
	}
	return errors.New("no dialogs matched provided label")
}