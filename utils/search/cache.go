package search

import (
	"errors"
	"github.com/infinit-lab/gravity/printer"
	"sync"
	"time"
)

type cache struct {
	buffer          []byte
	fh              frameHandler
	packages        map[uint16]map[uint16][]byte
	packagesTimeout map[uint16]int
	packagesMutex   sync.Mutex
	ticker          *time.Ticker
	quitChan        chan int
}

type frameHandler interface {
	onGetFrame(frame []byte)
}

const (
	frameHeadL         byte = 0x00
	frameHeadH         byte = 0x01
	frameCountL        byte = 0x02
	frameCountH        byte = 0x03
	framePackageIndexL byte = 0x04
	framePackageIndexH byte = 0x05
	frameIndexL        byte = 0x06
	frameIndexH        byte = 0x07
	frameDataLength    byte = 0x08
	frameData          byte = 0x09

	frameSingleDataLength byte = 0x04
	frameSingleData       byte = 0x05
)

func newCache(fh frameHandler) *cache {
	c := new(cache)
	c.fh = fh
	c.packages = make(map[uint16]map[uint16][]byte)
	c.packagesTimeout = make(map[uint16]int)
	c.ticker = time.NewTicker(time.Second)
	c.quitChan = make(chan int)
	go func() {
		for {
			isQuit := false
			select {
			case <-c.ticker.C:
				c.packagesMutex.Lock()
				for key, value := range c.packagesTimeout {
					if value > 30 {
						delete(c.packagesTimeout, key)
						delete(c.packages, key)
					}
					c.packagesTimeout[key] = value + 1
				}
				c.packagesMutex.Unlock()
			case <-c.quitChan:
				printer.Trace("quitChan")
				isQuit = true
			}
			if isQuit {
				c.ticker.Stop()
				break
			}
		}
	}()
	return c
}

func (c *cache) close() {
	close(c.quitChan)
}

func (c *cache) push(buffer []byte) {
	c.buffer = append(c.buffer, buffer...)
	for c.findHead() {
		frame, err := c.parseFrame()
		if err != nil {
			printer.Error("Failed to parseFrame. error: ", err)
			return
		}
		c.unpackFrame(frame)
	}
}

func (c *cache) findHead() bool {
	isFind := false
	index := 0
	for i, b := range c.buffer {
		if b == 0xAA {
			if i+1 < len(c.buffer) {
				if c.buffer[i+1] == 0xAA {
					isFind = true
					index = i
					break
				}
			} else {
				isFind = true
				index = i
				break
			}
		}
	}
	if isFind {
		c.buffer = c.buffer[index:]
	} else {
		c.buffer = []byte{}
	}
	return isFind
}

func byteToUint16(low, high byte) uint16 {
	return uint16(low) + uint16(high)<<8
}

func (c *cache) parseFrame() ([]byte, error) {
	if len(c.buffer) < int(frameCountH) {
		return nil, errors.New("buffer size is not enough")
	}
	frameCount := byteToUint16(c.buffer[frameCountL], c.buffer[frameCountH])
	dataLengthIndex := 0
	if frameCount == 1 {
		dataLengthIndex = int(frameSingleDataLength)
	} else {
		dataLengthIndex = int(frameDataLength)
	}
	if len(c.buffer) < dataLengthIndex {
		return nil, errors.New("buffer size is not enough")
	}
	dataLength := int(c.buffer[dataLengthIndex])
	if len(c.buffer)-dataLengthIndex+1 < dataLength {
		return nil, errors.New("buffer size is not enough")
	}
	frame := c.buffer[0 : dataLengthIndex+1+dataLength]
	c.buffer = c.buffer[dataLengthIndex+1+dataLength:]
	return frame, nil
}

func (c *cache) unpackFrame(frame []byte) {
	frameCount := byteToUint16(frame[frameCountL], frame[frameCountH])
	if frameCount == 1 {
		if c.fh != nil && len(frame) > int(frameSingleData) {
			c.fh.onGetFrame(frame[frameSingleData:])
		}
	} else {
		frameIndex := byteToUint16(frame[frameIndexL], frame[frameIndexH])
		packageIndex := byteToUint16(frame[framePackageIndexL], frame[framePackageIndexH])
		if c.packages != nil {
			c.packagesMutex.Lock()
			frames, ok := c.packages[frameIndex]
			if !ok {
				frames = make(map[uint16][]byte)
			}
			_, ok = c.packagesTimeout[frameIndex]
			if !ok {
				c.packagesTimeout[frameIndex] = 0
			}
			frames[packageIndex] = frame[frameData:]
			c.packages[frameIndex] = frames
			var buffer []byte
			if len(frames) == int(frameCount) {
				for i := uint16(0); i < frameCount; i++ {
					f, ok := frames[i]
					if !ok {
						buffer = []byte{}
						break
					}
					buffer = append(buffer, f...)
				}
				delete(c.packages, frameIndex)
				delete(c.packagesTimeout, frameIndex)
			}
			c.packagesMutex.Unlock()

			if len(buffer) != 0 {
				if c.fh != nil {
					c.fh.onGetFrame(buffer)
				}
			}
		} else {
			printer.Error("Packages is nil.")
		}
	}
}

func uint16ToByte(i uint16) []byte {
	var bytes []byte
	bytes = append(bytes, byte(i))
	bytes = append(bytes, byte(i>>8))
	return bytes
}

func packBuffer(buffer []byte, index uint16) [][]byte {
	var frameList [][]byte
	size := len(buffer)
	if size <= 0xFF {
		var frame []byte
		frame = append(frame, 0xAA, 0xAA, 0x01, 0x00, byte(size))
		frame = append(frame, buffer...)
		frameList = append(frameList, frame)
	} else {
		times := size / 0xFF
		if size%0xFF > 0 {
			times++
		}
		for i := 0; i < times; i++ {
			var temp []byte
			if i == times-1 {
				temp = append(temp, buffer[i*0xFF:]...)
			} else {
				temp = append(temp, buffer[i*0xFF:(i+1)*0xFF]...)
			}
			var frame []byte
			frame = append(frame, 0xAA, 0xAA)
			frame = append(frame, uint16ToByte(uint16(times))...)
			frame = append(frame, uint16ToByte(uint16(i))...)
			frame = append(frame, uint16ToByte(index)...)
			frame = append(frame, byte(len(temp)))
			frame = append(frame, temp...)
			frameList = append(frameList, frame)
		}
	}
	return frameList
}
