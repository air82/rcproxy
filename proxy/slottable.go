package proxy

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/collinmsn/resp"
	log "github.com/ngaut/logging"
)

const (
	NumSlots = 16384
)

var (
	ERR_INVALID_SLOT_INFO = errors.New("invalid slot info")
)

type SlotTable struct {
	slotServers []string
}

func NewSlotTable() *SlotTable {
	st := &SlotTable{
		slotServers: make([]string, NumSlots),
	}
	return st
}

func (st *SlotTable) Get(slot int) string {
	return st.slotServers[slot]
}

func (st *SlotTable) Set(slot int, server string) {
	st.slotServers[slot] = server
}

func (st *SlotTable) SetSlotInfo(si *SlotInfo) {
	for i := si.start; i <= si.end; i++ {
		st.Set(i, si.master)
	}
}

type SlotInfo struct {
	start  int
	end    int
	master string
}

func NewSlotInfo(data *resp.Data) (si *SlotInfo, err error) {
	if len(data.Array) < 3 || len(data.Array[2].Array) != 2 {
		log.Error(data.Array)
		return nil, ERR_INVALID_SLOT_INFO
	}
	host := string(data.Array[2].Array[0].String)
	if len(host) == 0 {
		host = "127.0.0.1"
	}
	si = &SlotInfo{
		start:  int(data.Array[0].Integer),
		end:    int(data.Array[1].Integer),
		master: fmt.Sprintf("%s:%d", host, int(data.Array[2].Array[1].Integer)),
	}
	return si, nil
}

func Key2Slot(key string) int {
	buf := []byte(key)
	if pos := bytes.IndexByte(buf, '{'); pos != -1 {
		pos += 1
		if pos2 := bytes.IndexByte(buf[pos:], '}'); pos2 > 0 {
			slot := CRC16(buf[pos:pos+pos2]) % NumSlots
			return int(slot)
		}
	}
	slot := CRC16(buf) % NumSlots
	return int(slot)
}
