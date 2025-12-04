package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aescanero/dago-libs/pkg/domain"
	"github.com/aescanero/dago-node-executor/internal/executor"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Worker represents an executor worker
type Worker struct {
	id           string
	redisClient  *redis.Client
	executor     *executor.Executor
	logger       *zap.Logger
	consumerGroup string
	streamKey    string

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	lastProcessed time.Time
	mu            sync.RWMutex
}

// Config holds worker configuration
type Config struct {
	ID           string
	RedisClient  *redis.Client
	Executor     *executor.Executor
	Logger       *zap.Logger
}

// NewWorker creates a new worker
func NewWorker(cfg *Config) *Worker {
	ctx, cancel := context.WithCancel(context.Background())

	return &Worker{
		id:            cfg.ID,
		redisClient:   cfg.RedisClient,
		executor:      cfg.Executor,
		logger:        cfg.Logger,
		consumerGroup: "executor-workers",
		streamKey:     "executor.work",
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start starts the worker
func (w *Worker) Start() error {
	w.logger.Info("starting worker", zap.String("worker_id", w.id))

	// Create consumer group if it doesn't exist
	err := w.redisClient.XGroupCreateMkStream(w.ctx, w.streamKey, w.consumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	w.wg.Add(1)
	go w.processLoop()

	w.logger.Info("worker started", zap.String("worker_id", w.id))
	return nil
}

// Stop stops the worker gracefully
func (w *Worker) Stop(ctx context.Context) error {
	w.logger.Info("stopping worker", zap.String("worker_id", w.id))

	w.cancel()

	// Wait for processing to finish with timeout
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		w.logger.Info("worker stopped gracefully", zap.String("worker_id", w.id))
		return nil
	case <-ctx.Done():
		return fmt.Errorf("worker stop timeout")
	}
}

// processLoop is the main worker loop
func (w *Worker) processLoop() {
	defer w.wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			w.processWork()
		}
	}
}

// processWork reads and processes work from Redis Stream
func (w *Worker) processWork() {
	// Read from stream
	streams, err := w.redisClient.XReadGroup(w.ctx, &redis.XReadGroupArgs{
		Group:    w.consumerGroup,
		Consumer: w.id,
		Streams:  []string{w.streamKey, ">"},
		Count:    1,
		Block:    time.Second,
	}).Result()

	if err != nil {
		if err == redis.Nil {
			// No messages, continue
			return
		}
		w.logger.Error("failed to read from stream", zap.Error(err))
		time.Sleep(time.Second)
		return
	}

	// Process messages
	for _, stream := range streams {
		for _, message := range stream.Messages {
			w.processMessage(message)
		}
	}
}

// processMessage processes a single work message
func (w *Worker) processMessage(message redis.XMessage) {
	w.logger.Info("processing work",
		zap.String("worker_id", w.id),
		zap.String("message_id", message.ID))

	// Parse work data
	workData, ok := message.Values["data"].(string)
	if !ok {
		w.logger.Error("invalid message format", zap.String("message_id", message.ID))
		w.ackMessage(message.ID)
		return
	}

	var work WorkItem
	if err := json.Unmarshal([]byte(workData), &work); err != nil {
		w.logger.Error("failed to unmarshal work", zap.Error(err))
		w.ackMessage(message.ID)
		return
	}

	// Execute node
	result, err := w.executeNode(&work)

	// Publish result
	if err != nil {
		w.publishResult(&work, nil, err)
	} else {
		w.publishResult(&work, result, nil)
	}

	// Acknowledge message
	w.ackMessage(message.ID)

	// Update last processed time
	w.mu.Lock()
	w.lastProcessed = time.Now()
	w.mu.Unlock()
}

// executeNode executes a node
func (w *Worker) executeNode(work *WorkItem) (interface{}, error) {
	// Load state from Redis
	state, err := w.loadState(work.GraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	// Create node config
	nodeConfig := &executor.NodeConfig{
		NodeID: work.NodeID,
		Config: work.Config,
	}

	// Execute
	result, err := w.executor.Execute(w.ctx, state, nodeConfig)
	if err != nil {
		return nil, err
	}

	// Update state
	if state.NodeStates == nil {
		state.NodeStates = make(map[string]*domain.NodeState)
	}

	nodeState := state.NodeStates[work.NodeID]
	if nodeState == nil {
		now := time.Now()
		nodeState = &domain.NodeState{
			NodeID:    work.NodeID,
			StartedAt: &now,
		}
		state.NodeStates[work.NodeID] = nodeState
	}

	now := time.Now()
	nodeState.Output = result
	nodeState.Status = domain.ExecutionStatusCompleted
	nodeState.CompletedAt = &now

	// Save updated state
	if err := w.saveState(state); err != nil {
		w.logger.Error("failed to save state", zap.Error(err))
	}

	return result, nil
}

// publishResult publishes execution result
func (w *Worker) publishResult(work *WorkItem, result interface{}, err error) {
	var eventType string
	var data map[string]interface{}

	if err != nil {
		eventType = "node.failed"
		data = map[string]interface{}{
			"graph_id": work.GraphID,
			"node_id":  work.NodeID,
			"error":    err.Error(),
		}
	} else {
		eventType = "node.completed"
		data = map[string]interface{}{
			"graph_id": work.GraphID,
			"node_id":  work.NodeID,
			"output":   result,
		}
	}

	event := map[string]interface{}{
		"id":        uuid.New().String(),
		"type":      eventType,
		"graph_id":  work.GraphID,
		"node_id":   work.NodeID,
		"timestamp": time.Now(),
		"data":      data,
	}

	eventJSON, _ := json.Marshal(event)

	streamKey := fmt.Sprintf("dago:events:%s", eventType)
	err = w.redisClient.XAdd(w.ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: map[string]interface{}{
			"data": string(eventJSON),
		},
	}).Err()

	if err != nil {
		w.logger.Error("failed to publish result", zap.Error(err))
	}
}

// ackMessage acknowledges a message
func (w *Worker) ackMessage(messageID string) {
	err := w.redisClient.XAck(w.ctx, w.streamKey, w.consumerGroup, messageID).Err()
	if err != nil {
		w.logger.Error("failed to ack message",
			zap.String("message_id", messageID),
			zap.Error(err))
	}
}

// loadState loads graph state from Redis
func (w *Worker) loadState(graphID string) (*domain.GraphState, error) {
	key := fmt.Sprintf("dago:state:%s", graphID)

	data, err := w.redisClient.Get(w.ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	var state domain.GraphState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// saveState saves graph state to Redis
func (w *Worker) saveState(state *domain.GraphState) error {
	key := fmt.Sprintf("dago:state:%s", state.GraphID)

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	err = w.redisClient.Set(w.ctx, key, data, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// WorkItem represents a work item from the stream
type WorkItem struct {
	GraphID      string                 `json:"graph_id"`
	NodeID       string                 `json:"node_id"`
	NodeType     domain.NodeType        `json:"node_type"`
	Config       map[string]interface{} `json:"config"`
	Dependencies []string               `json:"dependencies"`
}

// GetLastProcessed returns the last processed time
func (w *Worker) GetLastProcessed() time.Time {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.lastProcessed
}

// IsHealthy returns whether the worker is healthy
func (w *Worker) IsHealthy() bool {
	// Check if Redis is connected
	if err := w.redisClient.Ping(w.ctx).Err(); err != nil {
		return false
	}
	return true
}
