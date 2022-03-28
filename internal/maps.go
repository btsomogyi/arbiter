package internal

import "github.com/btsomogyi/arbiter/interfaces"

type messageMap struct {
	msgMap map[int64]message
}

func newMessageMap() *messageMap {
	return &messageMap{
		msgMap: make(map[int64]message),
	}
}

// getMessage returns the message stored at index key.
func (mm *messageMap) getMessage(key int64) (message, bool) {
	m, ok := mm.msgMap[key]
	return m, ok
}

// findMessage determines if the exact message is stored in messageMap.
func (mm *messageMap) containsMessage(m message) bool {
	key := m.request().GetKey()
	if message, found := mm.msgMap[key]; found {
		return message.same(m)
	}
	return false
}

// remove removes the message stored at index key.
func (mm *messageMap) remove(m message) {
	key := m.request().GetKey()
	delete(mm.msgMap, key)
}

// add the message to the messageMap.
func (mm *messageMap) add(m message) {
	key := m.request().GetKey()
	mm.msgMap[key] = m
}

// length returns the number of entries in the messageMap.
func (mm *messageMap) length() int {
	return len(mm.msgMap)
}

// Dump returns all key/values in messageMap for debugging purposes.
func (mm *messageMap) dump() []interfaces.Request {
	var dump []interfaces.Request
	for _, v := range mm.msgMap {
		dump = append(dump, v.request())
	}
	return dump
}
