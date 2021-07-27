package logrus_log_hook

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"net/http"
)

type (

	// Log服务端点配置.
	Hook struct {
		client   http.Client
		levels   []logrus.Level
		endpoint string
		name     string

		// 方法在执行请求之前执行
		BeforePost BeforeFunc

		// 方法在执行请求后执行
		AfterPost AfterFunc
	}

	BeforeFunc func(req *http.Request) error

	AfterFunc func(res *http.Response) error

	// Teamones Log日志格式
	Log struct {
		Level        string         `json:"level"`         // 日志级别：notice,warning,error
		Route        interface{}    `json:"route"`         // 请求地址, 可不传
		RequestParam datatypes.JSON `json:"request_param"` // 请求参数 json对象，可不传
		Record       string         `json:"record"`        // 日志内容
		BelongSystem string         `json:"belong_system"` // 所属系统，获取端点配置
	}

	RequestParamJson struct {
		Data logrus.Fields `json:"data"` // 日志级别：notice,warning,error
	}
)

// New创建钩子类型的实例，指定产生钩子的应用程序的名称
func New(name, endpoint string, levels []logrus.Level) *Hook {
	return &Hook{
		client:   http.Client{},
		levels:   levels,
		endpoint: endpoint,
		name:     name,
	}
}

// Levels返回由该钩子处理的所有级别的一个片段
func (h Hook) Levels() []logrus.Level {
	return h.levels
}

// Fire处理 HTTP POST请求
func (h Hook) Fire(entry *logrus.Entry) error {
	RequestParamJson, err := json.Marshal(entry.Data["request_param"])

	if err != nil {
		return err
	}

	log := Log{
		Level:        "error",
		Route:        entry.Data["route"],
		RequestParam: RequestParamJson,
		Record:       entry.Message,
		BelongSystem: h.name,
	}

	// 将日志数据转换为JSON
	payload, err := json.Marshal(log)

	if err != nil {
		return err
	}

	body := bytes.NewBuffer(payload)

	// 创建POST请求
	req, err := http.NewRequest("POST", h.endpoint, body)

	if err != nil {
		return err
	}

	// 设置适当的标头，以便我们可以识别服务
	req.Header.Add("service-name", h.name)
	req.Header.Add("content-type", "application/json")

	// 如果已经配置了自定义的before post处理程序，则运行该处理程序
	if h.BeforePost != nil {
		if err := h.BeforePost(req); err != nil {
			return err
		}
	}

	// 发送请求
	resp, err := h.client.Do(req)

	if err != nil {
		return err
	}

	// 如果已经配置了自定义后处理程序，则运行它
	if h.AfterPost != nil {
		if err := h.AfterPost(resp); err != nil {
			return err
		}
	}

	// 如果状态代码大于201，则返回错误
	if resp.StatusCode > http.StatusCreated {
		return errors.New(fmt.Sprintf("failed to post payload, the server responded with a status of %v", resp.StatusCode))
	}

	return nil
}
