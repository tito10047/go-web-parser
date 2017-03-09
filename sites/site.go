package sites

import (
	"stavkova/database"
	"reflect"
	"sync"
)

type Task struct {
	TaskNum int8
	Url string
	Params []interface{}
}

type TaskStack struct {
	taskLocker sync.Mutex
	taskCount int
	tasks chan *Task
}

func NewTaskStack(routinesCount int) *TaskStack {
	return &TaskStack{tasks:make(chan *Task, routinesCount)}
}

func (ts *TaskStack) AddTask(task int8, url string, args []interface{})  {
	ts.taskLocker.Lock()
	defer ts.taskLocker.Unlock()
	ts.taskCount++
	ts.tasks <- &Task{
		task,
		url,
		args,
	}
}

func (ts *TaskStack) EndTask()  {
	ts.taskLocker.Lock()
	defer ts.taskLocker.Unlock()
	ts.taskCount--
	if ts.taskCount==0 {
		close(ts.tasks)
	}
}

func (ts *TaskStack) NextTask() (*Task, bool) {
	task, ok := <- ts.tasks
	return task, ok
}

func (ts *TaskStack ) HasTask() bool {
	return ts.taskCount==0
}

func (ts *TaskStack) CloseTasks() {
	close(ts.tasks)
}

type Site interface {
	Setup(routinesCount int)
	ParseNext()
	HasNext() bool
	GetId() int
}
type NewSite func(id int, db *database.Database) *Site

var sites = map[string]NewSite{}

func RegisterSite(siteName string, constructor NewSite) {
	sites[siteName] = constructor
}

func NextSite(getId func(siteName string) (int, bool), db *database.Database) <-chan *Site {
	ch := make(chan *Site, 1)
	go func() {
		for name := range sites {
			siteId, ok := getId(name)
			if !ok {
				continue
			}
			site := createSite(name, siteId, db)
			ch <- site
		}
		close(ch)
	}()
	return ch
}

func createSite(name string, id int, db *database.Database) *Site {
	f := reflect.ValueOf(sites[name])

	in := []reflect.Value{
		reflect.ValueOf(id),
		reflect.ValueOf(db),
	}

	result := f.Call(in)
	//site := result[0].Convert(reflect.TypeOf((*Site)(nil)))
	site := result[0].Interface().(*Site)
	return site
}
