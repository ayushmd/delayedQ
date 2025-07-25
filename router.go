package cue

import (
	"regexp"
	"sync"

	"github.com/panjf2000/ants"
)

type Router struct {
	mu     sync.RWMutex
	queues map[string]*Queue
	pool   *ants.Pool
}

func NewRouter() *Router {
	p, _ := ants.NewPool(5000)
	// defer pool.Release()
	return &Router{
		queues: make(map[string]*Queue),
		pool:   p,
	}
}

func (r *Router) GetMatchingQueues(pattern string) []*Queue {
	r.mu.RLock()
	defer r.mu.RUnlock()
	arr := make([]*Queue, 0)
	for k, v := range r.queues {
		if k == pattern {
			arr = append(arr, v)
		} else if re := regexp.MustCompile(pattern); re.MatchString(k) {
			arr = append(arr, v)
		}
	}
	return arr
}

func (r *Router) Close() {
	r.pool.Release()
}

func (r *Router) CheckExsists(pattern string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for k := range r.queues {
		if k == pattern {
			return true
		}
	}
	return false
}

func (r *Router) CheckPatternExsists(pattern string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for k := range r.queues {
		if k == pattern {
			return true
		} else if re := regexp.MustCompile(pattern); re.MatchString(k) {
			return true
		}
	}
	return false
}

func (r *Router) initQueues(qs []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, q := range qs {
		r.queues[q] = NewQueue(q)
	}
}

func (r *Router) CreateQueue(qname string) error {
	if r.CheckExsists(qname) {
		return &QueueExsistsError{}
	}
	q := NewQueue(qname)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.queues[qname] = q
	return nil
}

func (r *Router) DeleteQueue(qname string) error {
	if !r.CheckExsists(qname) {
		return &QueueDoesNotExsists{}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.queues, qname)
	return nil
}

func (r *Router) ListQueues() []string {
	arr := make([]string, 0)
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, q := range r.queues {
		arr = append(arr, q.Name)
	}
	return arr
}

func (r *Router) AddListener(qname string, l Listener) error {
	r.mu.RLock()
	q, ok := r.queues[qname]
	if !ok {
		return &QueueDoesNotExsists{}
	}
	r.mu.RUnlock()
	r.mu.Lock()
	defer r.mu.Unlock()
	q.Listeners = append(q.Listeners, l)
	return nil
}

func (r *Router) RemoveListener(qname string, id int) {
	que := r.queues[qname]
	que.mu.Lock()
	defer que.mu.Unlock()
	var ind int = -1
	for i, v := range que.Listeners {
		if v.id == id {
			ind = i
			break
		}
	}
	if ind != -1 {
		que.Listeners = append(que.Listeners[:ind], que.Listeners[ind+1:]...)
	}
}

type NotifyRequest struct {
	Id   int64 `json:"id"`
	Data any   `json:"data"`
}

type SentItemResponse struct {
	sent      bool
	queueName string
}

func (r *Router) SendItem(item Item, ack bool) []SentItemResponse {
	arr := make([]SentItemResponse, 0)
	qs := r.GetMatchingQueues(item.QueueName)
	if len(qs) == 0 {
		return arr
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, q := range qs {
		if len(q.Listeners) == 0 {
			arr = append(arr, SentItemResponse{
				queueName: q.Name,
				sent:      false,
			})
		} else {
			q.ind = (q.ind + 1) % len(q.Listeners)
			l := q.Listeners[q.ind]
			_ = r.pool.Submit(func() {
				l.send(item.Id, item.Data, ack)
			})
		}
	}
	return arr
}
