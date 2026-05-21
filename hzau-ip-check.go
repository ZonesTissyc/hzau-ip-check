package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"hzau-ip-check/campus"
	"hzau-ip-check/ipquery"
)

// APIResponse 定义返回给前端的结构体
type APIResponse struct {
	IP        string `json:"ip"`
	Country   string `json:"country"`  // 示例：中国-湖北-武汉
	City      string `json:"city"`     // 仅供参考，数据库很久没更新了
	StudentID string `json:"username"` // 校园网用户id
	Error     string `json:"error,omitempty"`
}

// 生成简易 Trace ID 的函数
func generateTraceID() string {
	return fmt.Sprintf("REQ-%d-%04d", time.Now().UnixNano(), rand.Intn(10000))
}

var globalIPQuerier *ipquery.Querier

// ================== 限流器配置 ==================
type visitor struct {
	secLimiter *rate.Limiter
	minLimiter *rate.Limiter
	lastSeen   time.Time
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex // 保护 visitors map 的并发安全
)

func getVisitor(ip string) *visitor {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		// secLimiter: 每秒最多2次 (产生速率2/秒，桶容量2)
		// minLimiter: 每分钟最多10次 (产生速率10/60秒，桶容量10)
		v = &visitor{
			secLimiter: rate.NewLimiter(2, 2),
			minLimiter: rate.NewLimiter(rate.Every(time.Minute/10), 10),
			lastSeen:   time.Now(),
		}
		visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v
}

// cleanupVisitors 定期清理长时间未访问的 IP，防止内存泄漏
func cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

// 获取前端真实 IP
func getClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		log.Printf("xForwardedFor成功: %s\n", ips)
		return strings.TrimSpace(ips[0])
	} else {
		log.Printf("xForwardedFor失败")
	}
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return strings.TrimSpace(xRealIP)
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// rateLimitMiddleware 限流中间件
func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		v := getVisitor(ip)

		if !v.minLimiter.Allow() || !v.secLimiter.Allow() {
			log.Printf("[WARN] 触发限流拦截，来源IP: %s", ip)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "请求过于频繁，请稍后再试",
			})
			return
		}
		next(w, r)
	}
}

// 处理 HTTP 请求的 handler
func infoHandler(w http.ResponseWriter, r *http.Request) {

	// 生成唯一Trace ID
	traceID := generateTraceID()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	//  获取 IP
	ip := getClientIP(r)

	log.Printf("[%s] [START] 收到查询请求，来源IP: %s", traceID, ip)

	// 可以取消下面这行的注释进行硬编码测试：
	// ip = "218.199.160.20"

	respData := APIResponse{IP: ip}

	// 调用 ipquery 模块获取信息
	if globalIPQuerier != nil {
		if ipInfo, err := globalIPQuerier.Find(ip); err == nil {
			respData.Country = ipInfo.Country
			respData.City = ipInfo.City
			log.Printf("[%s] [INFO] IP归属地查询成功: %s %s", traceID, ipInfo.Country, ipInfo.City)
		} else {
			log.Printf("[%s] [WARN] IP归属地查询失败: %v", traceID, err)
		}
	}

	// campus 模块获取校园网id
	studentID, err := campus.FetchStudentID(ip)
	if err != nil {
		log.Printf("[%s] [ERROR] 校园网id失败: %v", traceID, err)
		respData.Error = fmt.Sprintf("获取id失败: %v", err)
		log.Printf("[%s] [END] 接口请求异常终止", traceID)
		json.NewEncoder(w).Encode(respData)
		return
	}
	log.Printf("[%s] [INFO] 接口调用成功，解析id: %s", traceID, studentID)
	respData.StudentID = studentID

	// 可在此处插入进一步查询代码

	// 返回成功组装的数据
	log.Printf("[%s] [END] 请求处理成功并返回数据", traceID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(respData)
}

func main() {
	// 初始化 IP 查询器
	globalIPQuerier = ipquery.NewQuerier("qqwry.dat")
	if globalIPQuerier == nil {
		log.Println("警告: qqwry.dat 初始化失败，IP归属地查询将不可用")
	}

	go cleanupVisitors()

	// 注册路由并启动服务
	http.HandleFunc("/api/ipcheck", rateLimitMiddleware(infoHandler))

	port := ":7891"
	log.Printf("服务已启动，正在监听端口 %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("服务意外终止: %v\n", err)
		// 终止可检测端口是否被占用
	}
}
