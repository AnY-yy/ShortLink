package bloomFilter

import (
	"errors"
	"math"
	"math/big"
	"sync"

	"github.com/twmb/murmur3"
)

type BloomFilter interface {
	AddBloomFilterElem(data []byte)
	IsExistData(data []byte) bool
	InitBloomFilter() error
}

type Repo interface {
	GetAllShortURL() ([]string, error)
}

type SBloomFilter struct {
	bitArray    *big.Int     // 位数组 存储0/1
	bitArrayNum uint         // 位数组长度
	hashNum     uint         // 哈希数量
	mu          sync.RWMutex // 读写锁

	repo Repo // 数据层接口
}

// NewBloomFilter 创建布隆过滤器接口器
// @param n 预期的元素数量
// @param p 预期的错误率
// @return BloomFilter
func NewBloomFilter(n uint, p float64, repo Repo) (BloomFilter, error) {
	if n <= 0 || p <= 0 || p > 1 {
		return nil, errors.New("调用布隆过滤器接口器传入参数错误")
	}
	// 计算最优数组长度
	bitArrayNum := getOptimalBitArrayNum(n, p)
	// 计算最优哈希数量
	hashNum := getOptimalHashNum(bitArrayNum, n)

	return &SBloomFilter{
		bitArray:    big.NewInt(0),
		bitArrayNum: bitArrayNum,
		hashNum:     hashNum,
		repo:        repo,
	}, nil
}

// getOptimalBitArrayNum 计算最优数组长度
// 依据公式: m = -n * ln(p) / (ln(2) * ln(2))  其中n为预期元素数量 p为预期错误率
func getOptimalBitArrayNum(n uint, p float64) uint {
	// 向上取整
	BitArrayNum := math.Ceil(-float64(n) * math.Log(p) / (math.Ln2 * math.Ln2))
	if BitArrayNum == 0 {
		BitArrayNum = 1
	}
	return uint(BitArrayNum)
}

// getOptimalHashNum 计算最优哈希数量
// 依据公式: k = m * ln(2) / n  其中m为最优数组长度 n为预期元素数量
func getOptimalHashNum(m, n uint) uint {
	// 四舍五入
	hashNum := math.Round(float64(m) * math.Ln2 / float64(n))
	if hashNum == 0 {
		hashNum = 1
	}
	return uint(hashNum)
}

// GetHashValues 获取当前元素的hashNum个哈希值
func (sbf *SBloomFilter) getHashValues(data []byte) []uint {
	positions := make([]uint, sbf.hashNum)

	// 使用双哈希寒素计算哈希值
	hash1, hash2 := murmur3.Sum128(data)

	// for循环生成hashNum个哈希值
	for i := uint(0); i < sbf.hashNum; i++ {
		combined := hash1 + hash2*uint64(i)
		// 取模得到数组索引
		position := uint(combined % uint64(sbf.bitArrayNum))
		positions = append(positions, position)
	}
	return positions
}

// AddBloomFilterElem 添加元素到布隆过滤器中
func (sbf *SBloomFilter) AddBloomFilterElem(data []byte) {
	sbf.mu.Lock()
	defer sbf.mu.Unlock()

	if len(data) == 0 {
		return
	}

	position := sbf.getHashValues(data)
	for _, pos := range position {
		sbf.bitArray.SetBit(sbf.bitArray, int(pos), 1)
	}
}

// IsExistData 判断元素是否在布隆过滤器中
// 返回false表示元素绝不存在
// 返回true表示元素可能存在
func (sbf *SBloomFilter) IsExistData(data []byte) bool {
	sbf.mu.RLock()
	defer sbf.mu.RUnlock()

	if len(data) == 0 {
		return false
	}

	position := sbf.getHashValues(data)
	for _, pos := range position {
		if sbf.bitArray.Bit(int(pos)) == 0 {
			return false
		}
	}
	return true
}

// InitBloomFilter
// 每次项目重新启动 都将数据库中的数据写入到布隆过滤器中
func (sbf *SBloomFilter) InitBloomFilter() error {
	shortURLs, err := sbf.repo.GetAllShortURL()
	if err != nil {
		return err
	}
	for _, url := range shortURLs {
		// fmt.Println("短码:", url)
		sbf.AddBloomFilterElem([]byte(url))
	}
	return nil
}
