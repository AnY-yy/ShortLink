package bloom

import (
	"fmt"
	"math"
	"math/big"
	"sync"

	"github.com/spaolacci/murmur3"
)

// SBloomFilter
// self-BloomFilter 自定义布隆过滤器结构体
type SBloomFilter struct {
	bitArray *big.Int     // 比特数组
	m        uint         // 数组大小
	k        uint         // 哈希函数数量
	mu       sync.RWMutex // 读写锁
}

// NewBloomFilter 创建布隆过滤器实例
// @param n 预计存入的元素数量
// @param p 期望的错误率
// @return *SBloomFilter
func NewBloomFilter(n uint, p float64) (*SBloomFilter, error) {
	if n == 0 || p <= 0 || p >= 1 {
		return nil, fmt.Errorf("布隆过滤器参数异常 初始化实例失败")
	}

	// 计算最优数组长度m
	m := getOptimalBitArrayLength(n, p)
	// 计算最优哈希函数数量k
	k := getOptimalHashFunction(n, m)

	return &SBloomFilter{
		bitArray: big.NewInt(0),
		m:        m,
		k:        k,
	}, nil
}

// getOptimalBitArrayLength 计算最优数组长度m
// @param n 预计存入的元素数量
// @param p 期望的错误率
// @return uint 最优数组长度m
// 依据公式: m = -n * ln(p) / (ln(2) * ln(2))
func getOptimalBitArrayLength(n uint, p float64) uint {
	m := math.Ceil(-float64(n) * math.Log(p) / (math.Ln2 * math.Ln2)) // 向上取整 避免低配位图
	if m == 0 {
		m = 1
	}
	return uint(m)
}

// getOptimalHashFunction 计算最优哈希函数数量k
// @param n 预计存入的元素数量
// @param m 最优数组长度m
// @return uint 最优哈希函数数量k
// 依据公式: k = m * ln(2) / n
func getOptimalHashFunction(n, m uint) uint {
	// 具体与math.Ceil需要进行比较一下
	k := math.Round(float64(m) * math.Ln2 / float64(n)) // 四舍五入 避免哈希函数数量为0
	if k < 1 {
		k = 1
	}
	return uint(k)
}

// getHashPositions 计算元素的i个哈希下标
// @param data 元素数据
// @return []uint 元素的i个哈希下标
func (sbf *SBloomFilter) getHashPositions(data []byte) []uint {
	positions := make([]uint, 0, sbf.k) // 预分配容量 避免重复扩容

	// 使用murmur3双哈希生成2个独立的哈希值
	hash1, hash2 := murmur3.Sum128(data)

	for i := uint(0); i < sbf.k; i++ {
		// 后续自定义哈希值生成函数时 可以增大扰动因子 i -> i * i 让哈希值分布更均匀 避免哈希冲突增加
		combined := hash1 + uint64(i)*hash2
		// 取模得到比特位下标
		position := uint(combined % uint64(sbf.m))
		positions = append(positions, position)
	}
	return positions
}

// AddBloomFilterElem 添加元素到布隆过滤器
// @param data 元素数据
func (sbf *SBloomFilter) AddBloomFilterElem(data []byte) {
	sbf.mu.Lock() // 加写锁 防止并发修改bitArray
	defer sbf.mu.Unlock()

	if len(data) == 0 { // 空数据保护
		return
	}

	// 后续拓展方向 记录每个数据的插入次数
	position := sbf.getHashPositions(data)
	for _, pos := range position {
		sbf.bitArray.SetBit(sbf.bitArray, int(pos), 1)
	}
}

// IsExistElem 判断元素是否存在
// @param data 元素数据
// @return bool 元素是否存在
// false代表元素绝对不存在
// true代表元素可能存在
func (sbf *SBloomFilter) IsExistElem(data []byte) bool {
	sbf.mu.RLock() // 加读锁 防止并发修改bitArray
	defer sbf.mu.RUnlock()

	if len(data) == 0 { // 空数据保护
		return false
	}

	positions := sbf.getHashPositions(data)
	for _, pos := range positions {
		if sbf.bitArray.Bit(int(pos)) == 0 {
			return false
		}
	}
	return true
}
