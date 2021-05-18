package model

// CollectionEntity Pivot model for entities in collection
type CollectionEntity struct {
	CollectionID int64
	EntityID     int64

	Collection *Collection
	Entity     *Entity
}
