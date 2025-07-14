package controllers

import (
	"sync"
	"time"
	"log"
)

type DualBuffer struct {
	Buffer1       [][]byte
	Buffer2       [][]byte
	Size1         int64
	Size2         int64
	CurrentActive int // 0 = Buffer1 is active, 1 = Buffer2 is active
	Mutex         sync.RWMutex
}

var TopicBuffers = make(map[string]*DualBuffer)
var topicBufferMutex sync.RWMutex

const BUFFER_SIZE = 10 * 1024 * 1024 // 10 MB
const BUFFER_ROTATION_INTERVAL = 10 // 10 secs

func InitTopic(topic string) {
	topicBufferMutex.Lock()
	defer topicBufferMutex.Unlock()

	if _, exists := TopicBuffers[topic]; !exists {
		TopicBuffers[topic] = &DualBuffer{
			Buffer1:       [][]byte{},
			Buffer2:       [][]byte{},
			Size1:         0,
			Size2:         0,
			CurrentActive: 0,
		}
		go startBufferRotation(TopicBuffers[topic], BUFFER_ROTATION_INTERVAL*time.Second)
	}
}

func AddMessage(topic string, message []byte) {
	topicBufferMutex.RLock()
	defer topicBufferMutex.RUnlock()

	if buffer, ok := TopicBuffers[topic]; ok {
		buffer.Mutex.Lock()
		defer buffer.Mutex.Unlock()

		if buffer.CurrentActive == 0 {
			log.Printf("[Topic: %s] Writing to Buffer1 | Current Size: %d bytes", topic, buffer.Size1)
			buffer.Buffer1 = append(buffer.Buffer1, message)
			buffer.Size1 += int64(len(message))
			log.Printf("[Topic: %s] Buffer1 new size: %d bytes", topic, buffer.Size1)

			if buffer.Size1 >= BUFFER_SIZE {
				log.Printf("[Topic: %s] Buffer1 exceeded size. Rotating to Buffer2", topic)
				buffer.Buffer2 = nil
				buffer.Size2 = 0
				buffer.CurrentActive = 1
			}
		} else {
			log.Printf("[Topic: %s] Writing to Buffer2 | Current Size: %d bytes", topic, buffer.Size2)
			buffer.Buffer2 = append(buffer.Buffer2, message)
			buffer.Size2 += int64(len(message))
			log.Printf("[Topic: %s] Buffer2 new size: %d bytes", topic, buffer.Size2)

			if buffer.Size2 >= BUFFER_SIZE {
				log.Printf("[Topic: %s] Buffer2 exceeded size. Rotating to Buffer1", topic)
				buffer.Buffer1 = nil
				buffer.Size1 = 0
				buffer.CurrentActive = 0
			}
		}
	}
}

func GetCurrentBuffer(topic string) [][]byte {
	topicBufferMutex.RLock()
	defer topicBufferMutex.RUnlock()

	if buffer, ok := TopicBuffers[topic]; ok {
		buffer.Mutex.RLock()
		defer buffer.Mutex.RUnlock()

		if buffer.CurrentActive == 0 {
			return buffer.Buffer1
		}
		return buffer.Buffer2
	}
	return nil
}

func startBufferRotation(buffer *DualBuffer, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		buffer.Mutex.Lock()
		if buffer.CurrentActive == 0 {
			buffer.Buffer2 = nil
			buffer.Size2 = 0
			buffer.CurrentActive = 1
		} else {
			buffer.Buffer1 = nil
			buffer.Size1 = 0
			buffer.CurrentActive = 0
		}
		buffer.Mutex.Unlock()
	}
}
