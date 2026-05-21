package campus

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// UserInfo 定义我们想要提取的字段结构体
type UserInfo struct {
	UserName string `json:"user_name"`
	Domain   string `json:"domain"`
}

// FetchStudentID 根据 IP 获取用户id
func FetchStudentID(targetIP string) (string, error) {
	baseURL := "http://rz.hzau.edu.cn/cgi-bin/rad_user_info"
	params := url.Values{}
	params.Set("callback", "jQuery112402268415882725554_1779070992470")
	params.Set("ip", targetIP)
	params.Set("_", "1779070992471")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := http.Get(fullURL)
	if err != nil {
		return "", fmt.Errorf("请求发起失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应内容失败: %v", err)
	}

	// 提取并解析 JSONP
	jsonpStr := string(body)
	startIdx := strings.IndexByte(jsonpStr, '(')
	endIdx := strings.LastIndexByte(jsonpStr, ')')

	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		return "", fmt.Errorf("无效的 JSONP 格式，找不到匹配的括号")
	}

	pureJsonStr := jsonpStr[startIdx+1 : endIdx]

	var info UserInfo
	err = json.Unmarshal([]byte(pureJsonStr), &info)
	if err != nil {
		return "", fmt.Errorf("JSON 解析失败: %v", err)
	}

	return info.UserName, nil
}
