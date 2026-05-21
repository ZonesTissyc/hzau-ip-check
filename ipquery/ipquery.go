package ipquery

import (
	"errors"
	"sync"

	"github.com/yinheli/qqwry"
)

// IPInfo 定义返回的 IP 信息结构体
type IPInfo struct {
	IP      string `json:"ip"`
	Country string `json:"country"`
	City    string `json:"city"`
}

type Querier struct {
	db *qqwry.QQwry
	mu sync.Mutex
}

func NewQuerier(dataPath string) *Querier {
	q := qqwry.NewQQwry(dataPath)

	return &Querier{
		db: q,
	}
}

func (q *Querier) Find(ip string) (IPInfo, error) {
	if q.db == nil {
		return IPInfo{}, errors.New("QQWry 数据库未初始化")
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	q.db.Find(ip)

	return IPInfo{
		IP:      q.db.Ip,
		Country: q.db.Country,
		City:    q.db.City,
	}, nil
}
