// Package event contains redis stream events
package event

const (
	// UPDATE redis action for add or update entry in cache
	UPDATE = "UPDATE"
	// DELETE redis action for delete from cache
	DELETE = "DELETE"
)
