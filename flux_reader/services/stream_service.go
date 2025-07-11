package services

import (
	"context"
	"fmt"
	"log"
	"time"
	"bytes"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/websocket"
)

var POLLING_WINDOW int = 1

// StartPingLoop sends periodic WebSocket pings and stops on failure
func StartPingLoop(conn *websocket.Conn, doneChan chan struct{}) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("Ping failed:", err)
				close(doneChan)  // Signal PollAndStream to stop
				return
			}
		case <-doneChan:
			return
		}
	}
}


// PollAndStreamMessages continuously polls S3 for new message files and streams them to the subscriber
func PollAndStreamMessages(conn *websocket.Conn, topic string, subscriberID string, checkpoint *SubscriberCheckpoint, doneChan <-chan struct{}, ackChan <-chan map[string]string) {
	ticker := time.NewTicker(time.Duration(POLLING_WINDOW) * time.Second) // Poll every 5 seconds (adjust as needed)
	defer ticker.Stop()

	log.Printf("Starting polling for subscriber [%s] on topic [%s]", subscriberID, topic)

	for {
		select {
			case <-doneChan:
				log.Printf("Polling Stopped for Subscriber [%s] on topic [%s]", subscriberID, topic)
				return

			case <-ticker.C:
				err := checkAndStreamNewMessages(conn, topic, subscriberID, checkpoint, ackChan)
				if err != nil {
					log.Printf("Streaming error for subscriber %s: %v", subscriberID, err)
					return // Exit on error (client likely disconnected)
				}
		}
		
	}
}

// checkAndStreamNewMessages fetches new files from S3, streams them, and updates checkpoint
func checkAndStreamNewMessages(conn *websocket.Conn, topic, subscriberID string, checkpoint *SubscriberCheckpoint, ackChan <-chan map[string]string) error {
	const maxRetries = 3
	prefix := fmt.Sprintf("rayflux/%s/", topic)

	log.Printf("Polling S3 for new files under prefix: %s", prefix)

	resp, err := S3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: &BucketName,
		Prefix: &prefix,
	})
	if err != nil {
		return fmt.Errorf("failed to list S3 files: %w", err)
	}

	for _, obj := range resp.Contents {
		fileKey := *obj.Key

		// Skip files already read or older
		if fileKey <= checkpoint.LastReadFile {
			continue
		}

		log.Printf("Prefetching new message file: %s", fileKey)

		content, err := downloadFileFromS3(fileKey)
		if err != nil {
			log.Printf("Failed to download %s: %v", fileKey, err)
			continue
		}

		log.Printf("Sending file %s to subscriber [%s] (%d bytes)", fileKey, subscriberID, len(content))


		message := map[string]interface{}{
			"batch_id": fileKey,
			"messages": string(content),
		}

		messageBytes, err:= json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}

		var success bool
		var retries int
		for retries=0; retries < maxRetries; retries++{

			log.Printf("Retry %d/3", retries+1)

			// send the msg batch
			err = conn.WriteMessage(websocket.TextMessage, messageBytes)
			if err != nil{
				log.Printf("Failed to send message: %v", err)
				break
			}

			select {
			case ack := <-ackChan:
				if ack["batch_id"] == fileKey && ack["status"] == "ok" {
					log.Printf("Ack Received: %s", ack["batch_id"])

					checkpoint.LastReadFile = fileKey
					checkpoint.LastReadTimeUTC = time.Now().UTC().Format(time.RFC3339)

					err = SaveCheckpoint(topic, subscriberID, checkpoint)
					if err != nil {
						log.Printf("Failed to save checkpoint: %v", err)
					} else {
						log.Printf("Checkpoint updated: LastReadFile = %s", fileKey)
					}

					success = true

					break
				} else {
					log.Printf("Invalid ACK received: %+v", ack)
					continue
				}
			case <-time.After(10 * time.Second):
				log.Printf("ACK timeout on attempt %d", retries+1)
				continue
			}

			if success{
				break
			}
		}

		

		if retries == maxRetries {
			log.Printf("No ACK received after %d retries. Closing connection.", maxRetries)

			closeErr:= conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "No ACK received"))
			if closeErr != nil {
				log.Printf("Error sending close message: %v", closeErr)
			}

			conn.Close()
			
			return fmt.Errorf("client unresponsive, terminating session")
		}
		
	}

	return nil
}

// downloadFileFromS3 reads the entire file from S3 and returns it as []byte
func downloadFileFromS3(key string) ([]byte, error) {
	log.Printf("⬇Downloading file from S3: %s", key)
	resp, err := S3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &BucketName,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("File %s downloaded (%d bytes)", key, buf.Len())
	return buf.Bytes(), nil
}
