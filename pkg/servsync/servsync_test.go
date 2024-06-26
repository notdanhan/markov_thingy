package servsync

import ("testing")

func TestServSyncGet(t *testing.T) {
	data := New("1234")

	mp := SyncMap{}

	mp.Set("123",data)

	data1, ok := mp.Get("123")
	if !ok {
		t.Fatal("Expected completion, got nothing")	
	}

	data1.ChanId = "12345"

	if data.ChanId != data1.ChanId {
		t.Fatalf("Channels not matching\ndata: %s\ndata1: %s",data.ChanId,data1.ChanId)
	}
}

func TestGettingInvalidKeyFails(t *testing.T) {
	testMap := SyncMap{}

	_, ok := testMap.Get("1111")
	if ok {
		t.Fatal("Should return false")
	}
}

func TestDeleteFromSyncMap(t *testing.T) {
	data := New("1234")

	mp := SyncMap{}

	mp.Set("123",data)

	if _, ok := mp.Get("123"); !ok {
		t.Fatal("Failed to add the ServSync struct to the map")
	}

	mp.Delete("123")

	if _, ok := mp.Get("123"); ok {
		t.Fatal("Failed to delete from map")
	}
}
