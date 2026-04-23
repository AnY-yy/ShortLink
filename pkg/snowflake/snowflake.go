package snowflake

import (
	"errors"
	"sync"
	"time"
)

const (
	workerIDBits   = 7  //  机器ID 7位  将机器ID扩大为128
	sequenceIDBits = 3  // 时钟序列 3位  将同一毫秒级内的序列扩大八倍
	sequenceBits   = 12 // 时钟序列 12位 0~4095 4096个序列

	// 每个ID的最大值
	maxWorkerID   = -1 ^ (-1 << workerIDBits)   // 127
	maxSequenceID = -1 ^ (-1 << sequenceIDBits) // 7
	maxSequence   = -1 ^ (-1 << sequenceBits)   // 4095

	// 左偏移量
	workerIDShift   = sequenceBits + sequenceIDBits                // 12+3=15
	sequenceIDShift = sequenceBits                                 // 12
	timestampShift  = sequenceBits + sequenceIDBits + workerIDBits // 12+3+7=22

	// 起始时间戳2025-01-01 00:00:00
	startTimestamp = 1735689600000
)

type SnowFlake struct {
	mu            sync.Mutex // 互斥锁 保护线程  零值就是可用的锁 不需要初始化即可使用
	workerID      int64      // 机器ID
	sequenceID    int64      // 时钟序列ID
	sequence      int64      // 序列
	lastTimestamp int64      // 上次生成的时间戳
}

// NewSnowFlake 初始化SnowFlake结构体实例 向包外暴露
func NewSnowFlake(workerID int64) (*SnowFlake, error) {
	if workerID < 0 || workerID > maxWorkerID {
		return nil, errors.New("workerID out of range")
	}
	return &SnowFlake{
		workerID:      workerID,
		sequenceID:    0,
		sequence:      0,
		lastTimestamp: 0,
	}, nil
}

// getCurrentTimeStamp 获取当前的时间戳
func (s *SnowFlake) getCurrentTimeStamp() int64 {
	return time.Now().UnixNano()/int64(time.Millisecond) - startTimestamp
}

// waitForNextMilli 等待进入下一个毫秒
func (s *SnowFlake) waitForNextMilli(lastTimestamp int64) int64 {
	timestamp := s.getCurrentTimeStamp()
	for timestamp <= lastTimestamp { // 当前时间戳小于等于上次生成的时间戳  一直获取当前时间戳 直到进入下一个毫秒
		timestamp = s.getCurrentTimeStamp()
	}
	return timestamp
}

// GenerateSnowFlakeID 生成雪花ID
func (s *SnowFlake) GenerateSnowFlakeID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	timestamp := s.getCurrentTimeStamp()

	// 判断时间戳与序列
	if timestamp != s.lastTimestamp {
		// 如果时间戳与上次生成的时间戳不同  重置序列与时钟序列
		s.sequence = 0
		s.sequenceID = 0
		s.lastTimestamp = timestamp
	} else if s.sequence < maxSequence { // 如果序列号在范围内 序列号+1
		s.sequence++
	} else if s.sequenceID < maxSequenceID { // 如果序列号达到最大值 时钟序列号在范围内 时钟序列号+1 序列号重置为0
		s.sequenceID++
		s.sequence = 0
	} else { // 序列与时钟序列都达到最大值 进入下一毫秒
		timestamp = s.waitForNextMilli(s.lastTimestamp)
		s.sequence = 0
		s.sequenceID = 0
		s.lastTimestamp = timestamp
	}

	// 生成雪花ID 将 时间戳+机器ID+时钟序列+序列号 合并为一个64位整数
	id := (timestamp << timestampShift) |
		(s.workerID << workerIDShift) |
		(s.sequenceID << sequenceIDShift) |
		s.sequence

	return id
}
