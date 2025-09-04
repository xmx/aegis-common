package transport

import "sync"

type Huber[K comparable] interface {
	// Get 通过 ID 获取 peer。
	Get(id K) Peer[K]

	// Del 通过删除节点，并返回删除的数据，nil 代表之前无数据。
	Del(id K) Peer[K]

	// Put 存放节点并返回原来的数据（如果存在的话）。
	Put(p Peer[K]) (old Peer[K])

	// PutIfAbsent 当此 ID 没有时才存放节点，并返回是否放入成功。
	PutIfAbsent(p Peer[K]) bool
}

func NewHub[K comparable](capacity int) Huber[K] {
	if capacity < 0 {
		capacity = 0
	}

	return &simpleHub[K]{
		peers: make(map[K]Peer[K], capacity),
	}
}

type simpleHub[K comparable] struct {
	mutex sync.RWMutex
	peers map[K]Peer[K]
}

func (sh *simpleHub[K]) Get(id K) Peer[K] {
	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	return sh.peers[id]
}

func (sh *simpleHub[K]) Del(id K) Peer[K] {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	old := sh.peers[id]
	delete(sh.peers, id)

	return old
}

func (sh *simpleHub[K]) Put(p Peer[K]) Peer[K] {
	id := p.ID()

	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	old := sh.peers[id]
	sh.peers[id] = p

	return old
}

func (sh *simpleHub[K]) PutIfAbsent(p Peer[K]) bool {
	id := p.ID()

	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	_, exists := sh.peers[id]
	if !exists {
		sh.peers[id] = p
	}

	return !exists
}
