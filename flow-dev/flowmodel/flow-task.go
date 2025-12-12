package flowmodel

import "github.com/sebarcode/codekit"

type TaskType string

const (
	TaskProcess TaskType = "PROCESS"
	TaskReview  TaskType = "REVIEW"
)

type Task struct {
	ID                   string
	Name                 string
	Instruction          string
	TaskType             TaskType
	Users                []TaskUser
	StopRequestIfFail    bool
	StopRequestIfSuccess bool
	Config               codekit.M
}

type Route struct {
	FromID string
	ToID   string
}

type Tasks []Task
type Routes []Route
