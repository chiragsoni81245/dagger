package models

import "errors"

var NoRowsAffectedError = errors.New("No rows affected")

var AlreadyInRunningState = errors.New("Already in running state")

