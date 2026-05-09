package snowflake

import (
	"errors"
	"sync"
	"time"
)

type SnowflakeGenerator interface {
	GenerateSnowFlakeID() int64
	getCurrentTime() int64
	waitForNextMillSecond(currentTime int64) int64
}

type SnowFlake struct {
	workID        int64      // 机器ID
	sequenceID    int64      // 时钟序列ID
	sequence      int64      // 序列
	lastTimeStamp int64      // 上次生成时间戳
	mu            sync.Mutex // 互斥锁
}

const (
	// 7位机器ID 3位时钟序列 12位序列ID
	workIDBits     = 7 // 机器ID位数
	sequenceIDBits = 3
	sequenceBits   = 12

	// 每个ID的最大值
	maxWorkID     = -1 ^ (-1 << workIDBits)     // 127
	maxSequenceID = -1 ^ (-1 << sequenceIDBits) // 7
	maxSequence   = -1 ^ (-1 << sequenceBits)   // 4095

	// 左偏移量
	sequenceShift   = 0
	sequenceIDShift = sequenceBits + sequenceShift
	workIDShift     = sequenceIDShift + sequenceIDBits
	timeStampShift  = workIDShift + workIDBits

	// 起始时间 2025-01-01 00:00:00
	startTime = 1735689600000
)

// NewSnowFlake
// 返回一个SnowflakeGenerator接口
func NewSnowFlake(workerID int64) (SnowflakeGenerator, error) {
	if workerID > maxWorkID || workerID <= 0 {
		return nil, errors.New("机器ID数值不合法")
	}
	return &SnowFlake{
		workID:        workerID,
		sequenceID:    0,
		sequence:      0,
		lastTimeStamp: 0,
	}, nil
}

// getCurrentTime 获取当前的时间戳
func (s *SnowFlake) getCurrentTime() int64 {
	return time.Now().UnixNano()/int64(time.Millisecond) - startTime
}

// waitForNextMillSecond 等待到下一个毫秒
func (s *SnowFlake) waitForNextMillSecond(currentTime int64) int64 {
	timeStamp := s.getCurrentTime()
	for timeStamp <= currentTime {
		timeStamp = s.getCurrentTime()
	}
	return timeStamp
}

// GenerateSnowFlakeID 生成全局唯一雪花ID
func (s *SnowFlake) GenerateSnowFlakeID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	timeStamp := s.getCurrentTime()

	// 判断时间序列
	if timeStamp != s.lastTimeStamp {
		// 如果时间序列不同 重置序列ID
		s.sequence = 0
		s.sequenceID = 0
		s.lastTimeStamp = timeStamp
	} else if s.sequence < maxSequence { // 先递增序列
		s.sequence++
	} else if s.sequenceID < maxSequenceID { // 然后递增序列ID 重置序列
		s.sequenceID++
		s.sequence = 0
	} else { // 数值达到当前时间序列下的max 等待进入下一毫秒
		timeStamp = s.waitForNextMillSecond(timeStamp)
		s.sequence = 0
		s.sequenceID = 0
		s.lastTimeStamp = timeStamp
	}

	// 拼装: 时间戳 机器ID 时间序列ID 时间序列
	id := (timeStamp << timeStampShift) |
		(s.workID << workIDShift) |
		(s.sequenceID << sequenceIDShift) |
		(s.sequence | sequenceShift)

	return id
}
