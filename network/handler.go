package network

import (
	"log"
	"strconv"
)

type IMessageHandle interface {
	DoMessageHandler(IRequest)
	AddRouter(uint32, IRouter)
	StartWorkerPool()
	SendMessageToTaskQueue(IRequest)
}

type MessageHandle struct {
	APIs           map[uint32]IRouter
	WorkerPoolSize uint32
	TaskQueue      []chan IRequest
}

func NewMessageHandle() *MessageHandle {
	return &MessageHandle{
		APIs:           make(map[uint32]IRouter),
		WorkerPoolSize: Setting.WorkerPoolSize,
		TaskQueue:      make([]chan IRequest, Setting.WorkerPoolSize),
	}
}

func (m *MessageHandle) DoMessageHandler(req IRequest) {
	handler, ok := m.APIs[req.GetMessageID()]
	if !ok {
		log.Println("api message id = ", req.GetMessageID(), " is not found")
		return
	}
	handler.BeforeHook(req)
	handler.Handle(req)
	handler.AfterHook(req)
}

func (m *MessageHandle) AddRouter(id uint32, router IRouter) {
	if _, ok := m.APIs[id]; ok {
		panic("repeated api, id = " + strconv.Itoa(int(id)))
	}
	m.APIs[id] = router
	log.Println("add api id = ", id)
}

func (m *MessageHandle) StartOneWorker(workerID int, taskQueue chan IRequest) {
	log.Println("worker id = ", workerID, " is started")
	for {
		select {
		case request := <-taskQueue:
			m.DoMessageHandler(request)
		}
	}
}

func (m *MessageHandle) StartWorkerPool() {
	for i := 0; i < int(m.WorkerPoolSize); i++ {
		m.TaskQueue[i] = make(chan IRequest, Setting.MaxWorkerTaskLen)
		go m.StartOneWorker(i, m.TaskQueue[i])
	}
}

func (m *MessageHandle) SendMessageToTaskQueue(req IRequest) {
	workerID := req.GetConnection().GetConnID() % m.WorkerPoolSize
	log.Println("add connID = ", req.GetConnection().GetConnID(), " request messageID = ", req.GetMessageID(), " to workerID = ", workerID)
	m.TaskQueue[workerID] <- req
}
