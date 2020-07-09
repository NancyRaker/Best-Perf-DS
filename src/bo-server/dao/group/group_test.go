package group

import (
	"bo-server/engine/db"
	"testing"
)

func TestRemoveGroup(t *testing.T) {
	err := RemoveGroup(755, db.DBHandler)
	t.Log(err)
}