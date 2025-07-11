package controllers

import (
	"bytes"
	"flux_emitter/services"
	"fmt"
	"sync"
	"time"
	"encoding/json"
)

type BufferController struct{
	messageCaches map[string]*bytes.Buffer
	cacheSizes map[string]int
	lastFlushTimes map[string]time.Time
	mutex sync.Mutex
}

type StoredMessage struct {
	Topic string `json:"topic"`
	Timestamp string `json:"timestamp"`
	Message string `json:"message"`
}

const MaxCacheSize = 100 // 100 Bytes

func NewBufferController() *BufferController{
	return &BufferController{
		messageCaches: make(map[string]*bytes.Buffer),
		cacheSizes: make(map[string]int),
		lastFlushTimes: make(map[string]time.Time),
	}
}

func (bc *BufferController) AddMessage(topic string, message []byte) error{
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if bc.messageCaches[topic] == nil {
		bc.messageCaches[topic] = &bytes.Buffer{}
	}

	msg := StoredMessage{
		Topic: topic,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Message: string(message),
	}
	
	jsonBytes, err:= json.Marshal(msg)
	if err!=nil{
		return fmt.Errorf("Failed to marshal message: %w", err)
	}

	newSize := bc.cacheSizes[topic] + len(jsonBytes) + 1
	if newSize >= MaxCacheSize {
		_ = bc.flush(topic)
	}

	topicCache := bc.messageCaches[topic]
	topicCache.Write(jsonBytes)
	topicCache.Write([]byte("\n"))

	bc.cacheSizes[topic] += len(jsonBytes) + 1

	fmt.Printf("Current cache for topic [%s]: %d bytes\n", topic, bc.cacheSizes[topic])

	if bc.cacheSizes[topic] >= MaxCacheSize {
		return bc.flush(topic)
	}

	return nil
}

func (bc *BufferController) flush(topic string) error {
	topicCache := bc.messageCaches[topic]
	if topicCache == nil || topicCache.Len() == 0 {
		return nil
	}

	// Flush to S3
	err := services.FlushToS3(topic, topicCache.Bytes())
	if err != nil {
		return fmt.Errorf("failed to flush to S3: %w", err)
	}

	// Reset Cache and Size
	bc.messageCaches[topic] = &bytes.Buffer{}
	bc.cacheSizes[topic] = 0
	bc.lastFlushTimes[topic] = time.Now()

	return nil
}

func (bc *BufferController) flushAll(interval time.Duration) {
	bc.mutex.Lock()

	defer bc.mutex.Unlock()


	for topic, buffer := range bc.messageCaches {
		lastFlush, exists:=bc.lastFlushTimes[topic]
		if !exists || time.Since(lastFlush) >= interval {
			bytesFlushed := bc.cacheSizes[topic]

			if buffer == nil || buffer.Len() == 0{
				continue
			}

			if err:=bc.flush(topic); err==nil{
				fmt.Printf("[%s] Flushed %d bytes for topic [%s]\n", time.Now().Format(time.RFC3339), bytesFlushed, topic)
			}else {
				fmt.Printf("[%s] Failed to flush topic [%s]: %v\n", time.Now().Format(time.RFC3339), topic, err)
			}
		}
	}
}

func (bc *BufferController) StartAutoFlush(interval time.Duration){
	ticker := time.NewTicker(interval)

	go func(){
		for range ticker.C{
			bc.flushAll(interval)
		}
	}()
}