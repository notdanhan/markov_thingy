package servsync

import (
	"encoding/json"
	"sync"
)

// This is a wrapper for a sync map to allow for JSON serialization/Deserialization
// And to avoid typecasting in the application space

type SyncMap struct {
	smap sync.Map // The Map in question :)
}

// Get value from map
func (u *SyncMap) Get(key string) (val *ServSync, ok bool) {
	val = nil
	val1, ok := u.smap.Load(key)
	if !ok {
		return
	}
	val, ok = val1.(*ServSync)
	return
}

// Set value in map
func (u *SyncMap) Set(key string, val *ServSync) {
	u.smap.Store(key, val)
}

// Delete value from map
func (u *SyncMap) Delete(key string) {
	u.smap.Delete(key)
}

func (u *SyncMap) MarshalJSON() ([]byte, error) {
	var sMap map[string]ServSync

	u.smap.Range(func(key, value any) bool {
		if key == nil {
			return false
		}

		mKey := key.(string)
		mValue := value.(*ServSync)

		sMap[mKey] = *mValue

		return true
	})

	return json.Marshal(sMap)
}

func (u *SyncMap) UnmarshalJSON(data []byte) error {
	var sMap map[string]ServSync

	if err := json.Unmarshal(data, &sMap); err != nil {
		return err
	}

	u.smap = sync.Map{}

	for key, value := range sMap {
		u.smap.Store(key, value)
	}

	return nil
}
