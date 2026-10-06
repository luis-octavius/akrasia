package tasks 

import (
	"uuid"
	"context"	
	"log"

	database "github.com/luis-octavius/akrasia/internal/db/out"
)

var (
	ctx = context.Background();
)

func (tkm *TaskManager) AddDump(name string) error {
	dumpId := uuid.New();

	_, err := tkm.Queries.AddDump(ctx, database.AddDumpParams{
		ID: dumpId, 
		Name: name,
	})

	if err != nil {
		log.Fatalf("Error by adding dump with %v", name)
	}

	return nil
}

func (tkm *TaskManager) GetAllDump() ([]database.Dump, error) {
	allDump, err := tkm.Queries.GetAllDump(ctx)
	if err != nil {
		log.Fatal("Error getting all dump")
	}

	return allDump, nil
}
