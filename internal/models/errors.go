package models

import "errors"

var NoRowsAffectedError = errors.New("No rows affected")

var AlreadyInRunningState = errors.New("Already in running state")

var TaskIsStillRunning = errors.New("Task is still in running state")

var TaskNotStarted = errors.New("Task has not yet started")
