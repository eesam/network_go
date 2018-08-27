package logic

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const mcPkgHeaderLength = 5

var partMcData = errors.New("part mc data")

// 解析来自atc的数据
type decoder struct {
	dataBuf []byte
}

// Find Packet Flag: 0xac
func (p *decoder) findToHeaderFlag() {
	iOffset := 0
	for ; iOffset+3 < len(p.dataBuf); iOffset++ {
		if p.dataBuf[iOffset] == 0xac {
			break
		}
	}
	p.dataBuf = p.dataBuf[iOffset:]
}

func (p *decoder) bytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int16
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return int(tmp)
}

func (p *decoder) parseHeader() (string, error) {
	offset := 1
	length := p.bytesToInt(p.dataBuf[offset : offset+2])
	offset += 2

	reserved := p.bytesToInt(p.dataBuf[offset : offset+2])
	if reserved != 1 {
		gLogger.Debug("reserved is not 1")
		return "", errors.New("invalid mc data")
	}
	offset += 2

	if length+mcPkgHeaderLength > len(p.dataBuf) {
		return string(p.dataBuf[offset:]), partMcData
	}
	return string(p.dataBuf[offset : offset+length]), nil
}

func (p *decoder) decode(data []byte) []string {
	var dataLst []string
	p.dataBuf = append(p.dataBuf, data...)
	for len(p.dataBuf) > mcPkgHeaderLength {
		p.findToHeaderFlag()

		if len(p.dataBuf) <= mcPkgHeaderLength {
			break
		}

		jsonData, err := p.parseHeader()
		if err == partMcData {
			break
		} else if err != nil {
			p.dataBuf = p.dataBuf[1:]
			continue
		}

		dataLst = append(dataLst, jsonData)
		skipLen := mcPkgHeaderLength + len(jsonData)
		p.dataBuf = p.dataBuf[skipLen:]
	}
	return dataLst
}
