package topic

import (
	"github.com/go-pg/pg/v9"
	"github.com/refto/server/database"
	"github.com/refto/server/database/model"
	"github.com/refto/server/server/request"
)

func Filter(req request.FilterTopics) (data []model.Topic, count int, err error) {
	q := database.ORM().
		Model(&data)

	if req.Name != "" {
		q.Where("name ILIKE ?", req.Name+"%")
	}

	count, err = q.SelectAndCount()
	return
}

// Common return topics that in common with entities from given topics
// For example:
// 		entity1 have topics A, B, C
// 		entity2 have topics A, D, E
// For given topic A, should return B, C, D, E
// For given topic B, should return A, C
// For given topics A,B, should return C
// For given topics A,E, should return D
// (Also collection and repo should filtered if provided)
// TODO topics from hidden and private repos will leak!
func Common(in CommonTopicsParams) (out []string, err error) {
	// just return all topics
	if len(in.Topics) == 0 && in.CollectionID == 0 && in.RepoID == 0 {
		err = database.ORM().
			Model(&[]model.Topic{}).
			Column("name").
			Order("name").
			Select(&out)
		return
	}

	// just return all topics of collection
	if in.CollectionID != 0 && len(in.Topics) == 0 && in.RepoID == 0 {
		err = database.ORM().
			Model(&[]model.Topic{}).
			Column("name").
			Join("JOIN entity_topics et ON et.topic_id=topic.id").
			Join("JOIN collection_entities ce ON ce.entity_id=et.entity_id").
			Where("ce.collection_id = ?", in.CollectionID).
			Order("name").
			Group("topic.id").
			Select(&out)
		return
	}

	// just return all topics of repo
	if in.RepoID != 0 && len(in.Topics) == 0 && in.CollectionID == 0 {
		err = database.ORM().
			Model(&[]model.Topic{}).
			Where("repo_id = ?", in.RepoID).
			Column("name").
			Order("name").
			Select(&out)
		return
	}

	// select entities that having given topics
	entities := database.ORM().
		Model(&[]model.Topic{}).
		Column("e.id").
		Join("JOIN entity_topics et ON et.topic_id=topic.id").
		Join("JOIN entities e ON et.entity_id=e.id").
		WhereIn("topic.name IN (?)", in).
		Having("COUNT(topic.id) = ?", len(in.Topics)).
		Group("e.id")

	if in.CollectionID != 0 {
		entities = entities.
			Join("JOIN collection_entities ce ON ce.entity_id=e.id").
			Where("ce.collection_id = ?", in.CollectionID)
	}
	if in.RepoID != 0 {
		entities = entities.
			Where("topic.repo_id = ?", in.RepoID)
	}

	// select topics (of selected entities) that not a given topics
	err = database.ORM().
		Model().
		Column("t.name").
		With("selected_entities", entities).
		TableExpr("selected_entities se").
		Join("JOIN entity_topics et ON et.entity_id=se.id").
		Join("JOIN topics t ON et.topic_id=t.id").
		WhereIn("t.name NOT IN (?)", in).
		Order("name").
		Group("t.id").
		Select(&out)

	return
}

func FirstOrCreate(elem *model.Topic) (err error) {
	err = database.ORM().
		Model(&elem).
		Where("name = ?", elem.Name).
		Where("repo_id = ?", elem.RepoID).
		First()
	if err != nil && err != pg.ErrNoRows {
		return
	}

	if err == pg.ErrNoRows {
		err = database.ORM().Insert(&elem)
	}

	return
}
