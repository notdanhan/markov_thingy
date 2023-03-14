package servsync

import (
	"encoding/json"
	"sync/atomic"

	"github.com/danielh2942/markov_thingy/pkg/markovcommon"
	"github.com/google/uuid"
)

// Preliminary setup to allow for multi server support

type ServSync struct {
	ChanId      string                   // Channel that messages are read from
	FileName    string                   // name of file database is written to
	MsgCount    atomic.Int32             // count of messages sent
	MarkovChain markovcommon.MarkovChain // markov chain stored/used
}

func New(ChanId string) *ServSync {
	mUUID := uuid.New()
	return &ServSync{
		ChanId,
		mUUID.String() + ".json",
		atomic.Int32{},
		&markovcommon.MarkovData{
			StartWords: []uint{},
			WordCount:  0,
			WordRef:    map[string]uint{},
			WordVals:   []string{},
			WordGraph:  []map[uint]uint{},
		},
	}
}

func (u *ServSync) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ChanId   string `json:ChanId`
		FileName string `json:FileName`
	}{
		ChanId:   u.ChanId,
		FileName: u.FileName,
	})
}

func (u *ServSync) UnmarshalJSON(data []byte) error {
	aux := &struct {
		ChanId   string `json:ChanId`
		FileName string `json:FileName`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	u.FileName = aux.FileName
	u.MsgCount.Store(0)
	if tmp, err := markovcommon.ReadinFile(u.FileName); err != nil {
		return err
	} else {
		u.MarkovChain = tmp
	}
	return nil
}
