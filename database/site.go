package database

type DbSite struct {
	Id            int
	RoutinesCount int
	TasksPerTime  int
	WaitSeconds   int
	Name          string
}
