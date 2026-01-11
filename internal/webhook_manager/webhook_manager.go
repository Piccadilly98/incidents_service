package webhook_manager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Piccadilly98/incidents_service/internal/config"
	"github.com/Piccadilly98/incidents_service/internal/handlers"
	"github.com/Piccadilly98/incidents_service/internal/models/dto"
	"github.com/Piccadilly98/incidents_service/internal/repository"
)

const (
	DefaultURL            = "http://localhost:9090"
	DefaultMethod         = http.MethodPost
	DefaultMaxReTry       = 3
	DefaultTimeOutRequest = time.Second * 5

	PrefixRetryableError    = "retryable"
	PrefixNonRetryableError = "non-retryable"
)

type WebhookManager struct {
	defaultUrl    string
	defaultMethod string
	maxReTry      int
	cacheQueue    repository.CacheQueue
	webhookLogger *log.Logger
	ctx           context.Context
	cancel        context.CancelFunc
	httpClient    *http.Client
	backoff       bool
}

func NewWebhookManager(cfg *config.Config, cacheQueue repository.CacheQueue, maxReTry int, backoff bool, ctx context.Context) (*WebhookManager, error) {
	if cacheQueue == nil {
		return nil, fmt.Errorf("cacheQueue cannot be nil")
	}
	if maxReTry <= 0 {
		log.Printf("invalid maxReTry in parametrs, change to default: %d\n", DefaultMaxReTry)
		maxReTry = DefaultMaxReTry
	}
	cfgMethod := strings.ToUpper(cfg.WebhookMethod)
	cfgURL := cfg.WebhookURL

	if cfgMethod == "" || (cfgMethod != http.MethodPost && cfgMethod != http.MethodGet) {
		log.Printf("invalid webhook method in config, change to default: %s\n", DefaultMethod)
		cfgMethod = DefaultMethod
	}
	if cfgURL == "" {
		cfgURL = DefaultURL
		log.Printf("empty webhook url in config, change to default: %s\n", DefaultURL)
	}
	ctxRes, cancel := context.WithCancel(ctx)
	wm := &WebhookManager{
		cacheQueue:    cacheQueue,
		defaultUrl:    cfgURL,
		defaultMethod: cfgMethod,
		maxReTry:      maxReTry,
		webhookLogger: log.New(os.Stderr, "[WEBHOOK MANAGER]  ", log.Ldate|log.Ltime),
		ctx:           ctxRes,
		cancel:        cancel,
		httpClient: &http.Client{
			Timeout: DefaultTimeOutRequest,
		},
		backoff: backoff,
	}
	go wm.StartProcessing()
	return wm, nil
}

func (wm *WebhookManager) Stop() {
	wm.cancel()
}

func (wm *WebhookManager) AddToQueue(result dto.LocationCheckResponse, ctx context.Context, url, method string) error {
	if !result.IsDanger {
		return fmt.Errorf("invalid input: is_danger cannot be false")
	}
	if url == "" {
		url = wm.defaultUrl
	}
	if method == "" || (method != http.MethodPost && method != http.MethodGet) {
		method = wm.defaultMethod
	}
	body := &dto.WebhookTask{
		Dto:    result,
		Url:    url,
		Method: method,
	}

	err := wm.cacheQueue.AddToQueue(body, ctx)
	return err
}

func (wm *WebhookManager) StartProcessing() {
	for {
		select {
		case <-wm.ctx.Done():
			wm.webhookLogger.Println("CONTEX CANCEL, FINISH WORK")
		default:
			task, exists, err := wm.cacheQueue.PopFromQueue(wm.ctx)
			if err != nil {
				wm.webhookLogger.Printf("error in brpop: %s\n", err.Error())
				time.Sleep(500 * time.Millisecond)
				continue
			}

			if !exists {
				time.Sleep(300 * time.Millisecond)
				continue
			}
			err = wm.sendingRequest(task)
			if err != nil {
				if strings.Contains(err.Error(), PrefixRetryableError) {
					task.CountReTry++
					if task.CountReTry > wm.maxReTry {
						wm.webhookLogger.Printf("max retries exceeded for check_id=%s", task.Dto.ID)
						continue
					}
					if wm.backoff {
						delay := min(time.Duration(task.CountReTry*2)*time.Second, 30*time.Second)
						time.Sleep(delay)
					}
					err = wm.cacheQueue.PushTask(task, wm.ctx)
					if err != nil {
						wm.webhookLogger.Printf("error in re-push task: %s\nDelete task", err.Error())
					}
				} else {
					wm.webhookLogger.Printf("error in send request with check_id: %s, err: %s\nDelete task", task.Dto.ID, err.Error())
				}
				continue
			}
		}
	}
}

func (wm *WebhookManager) sendingRequest(task *dto.WebhookTask) error {
	var req *http.Request
	var err error

	if task.Method == http.MethodGet {
		req, err = http.NewRequestWithContext(wm.ctx, task.Method, task.Url, nil)
	} else {
		resultDto := task.ToResultWebhookDto()
		b, err := json.Marshal(resultDto)
		if err != nil {
			return fmt.Errorf("error in marshaling to request dto: %s\n", err.Error())
		}
		req, err = http.NewRequestWithContext(wm.ctx, task.Method, task.Url, bytes.NewBuffer(b))
		req.Header.Set(handlers.HeaderContentType, handlers.HeaderJson)
	}
	if err != nil {
		return fmt.Errorf("failed to create new request, err: %s\n", err.Error())
	}
	result, err := wm.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error in request: %s\n", err.Error())
	}
	defer result.Body.Close()

	_, _ = io.ReadAll(result.Body)

	if result.StatusCode >= 500 || result.StatusCode == 429 {
		return fmt.Errorf("retryable error: status %d", result.StatusCode)
	}
	if result.StatusCode >= 300 {
		return fmt.Errorf("non-retryable error: status %d", result.StatusCode)
	}

	return nil
}
