package sites

import (
	"stavkova/database"
	"reflect"
	"sync"
	"time"
)

const tasksClosed int = -1

type myTask struct {
	TaskNum int8
	Url     string
	Params  []interface{}
}

type TaskStack struct {
	totalTaskLocker   sync.Mutex
	taskCount         int
	tasks             chan *myTask
	currentTaskLocker sync.Mutex
	tasksPerTime      int
	startedTasks      int
	waitSeconds       int
}

func NewTaskStack(routinesCount, tasksPerTime, waitSeconds int) *TaskStack {
	t := &TaskStack{
		tasks:        make(chan *myTask),
		tasksPerTime: tasksPerTime,
		waitSeconds:  waitSeconds,
	}
	return t
}

func (ts *TaskStack) AddTask(task int8, url string, args []interface{}) {
	ts.totalTaskLocker.Lock()
	defer ts.totalTaskLocker.Unlock()
	if ts.taskCount==tasksClosed{
		return
	}
	ts.taskCount++
	go func(task int8, url string, args []interface{}) {
		ts.tasks <- &myTask{
			task,
			url,
			args,
		}
	}(task,url,args)
}

func (ts *TaskStack) closeTaskChannel(isLocked bool)  {
	if !isLocked {
		ts.totalTaskLocker.Lock()
		defer ts.totalTaskLocker.Unlock()
	}
	for ;ts.taskCount!=tasksClosed && ts.taskCount>0;ts.taskCount--{
		<-ts.tasks
	}
	ts.taskCount=tasksClosed
	close(ts.tasks)
}

func (ts *TaskStack) EndTask() {
	ts.totalTaskLocker.Lock()
	defer ts.totalTaskLocker.Unlock()
	ts.taskCount--
	if ts.taskCount == 0 {
		ts.closeTaskChannel(true)
	}
}

func (ts *TaskStack) NextTask() (*myTask, bool) {
	ts.currentTaskLocker.Lock()

	ts.startedTasks++
	if ts.startedTasks >= ts.tasksPerTime {
		ts.startedTasks = 0
		time.Sleep(time.Duration(ts.waitSeconds) * time.Second)
	}
	ts.currentTaskLocker.Unlock()

	task, ok := <-ts.tasks
	return task, ok
}

func (ts *TaskStack) HasTask() bool {
	ts.totalTaskLocker.Lock()
	defer ts.totalTaskLocker.Unlock()
	return ts.taskCount != tasksClosed
}

func (ts *TaskStack) CloseTasks() {
	ts.closeTaskChannel(false)
}

type SiteInt interface {
	Setup(routinesCount, tasksPerTime, waitSeconds int)
	ParseNext()
	HasNext() bool
	GetArgs() *database.DbSite
}
type NewSite func(dbSite *database.DbSite, db *database.Database) *SiteInt

var sites = map[string]NewSite{}

func RegisterSite(siteName string, constructor NewSite) {
	sites[siteName] = constructor
}

func NextSite(getId func(siteName string) (*database.DbSite, bool), db *database.Database) <-chan *SiteInt {
	ch := make(chan *SiteInt, 1)
	go func() {
		for name := range sites {
			dbSite, ok := getId(name)
			if !ok {
				continue
			}
			if dbSite.Enabled {
				site := createSite(name, dbSite, db)
				ch <- site
			}
		}
		close(ch)
	}()
	return ch
}

func createSite(name string, dbSite *database.DbSite, db *database.Database) *SiteInt {
	f := reflect.ValueOf(sites[name])

	in := []reflect.Value{
		reflect.ValueOf(dbSite),
		reflect.ValueOf(db),
	}

	result := f.Call(in)
	//site := result[0].Convert(reflect.TypeOf((*SiteInt)(nil)))
	site := result[0].Interface().(*SiteInt)
	return site
}
