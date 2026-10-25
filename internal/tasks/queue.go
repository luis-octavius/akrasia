package tasks

import "context"

// QueueSeen reports whether the offline-queue request id was already
// applied.
func (tkm *TaskManager) QueueSeen(ctx context.Context, id string) (bool, error) {
	seen, err := tkm.Queries.QueueSeen(ctx, id)
	return seen == 1, err
}

// QueueRemember records that the offline-queue request id was applied, and
// forgets the ones that are too old to be replayed.
func (tkm *TaskManager) QueueRemember(ctx context.Context, id string) error {
	if err := tkm.Queries.QueueRemember(ctx, id); err != nil {
		return err
	}
	return tkm.Queries.QueuePrune(ctx)
}
