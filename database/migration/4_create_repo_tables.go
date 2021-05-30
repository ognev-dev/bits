package migration

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	mg := `
CREATE TABLE repositories
(
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT      NOT NULL REFERENCES users (id),
    path          TEXT UNIQUE NOT NULL,
    name          TEXT        NOT NULL,
    description   TEXT,
    secret        TEXT        NOT NULL,
    type          TEXT        NOT NULL,
    confirmed     BOOL DEFAULT FALSE,
    clone_url     TEXT,
    import_status TEXT,
    import_log    TEXT,
    created_at    TIMESTAMPTZ NOT NULL,
    updated_at    TIMESTAMPTZ,
    import_at     TIMESTAMPTZ
);

CREATE TABLE user_repositories
(
    repo_id BIGINT NOT NULL REFERENCES repositories (id),
    user_id BIGINT NOT NULL REFERENCES users (id)
);

ALTER TABLE entities
    ADD COLUMN repo_id BIGINT REFERENCES repositories (id);

ALTER TABLE entities
    ALTER COLUMN repo_id SET NOT NULL;

DROP INDEX topics_name_key;
ALTER TABLE topics
    ADD COLUMN repo_id BIGINT REFERENCES repositories (id);
`

	migrations.MustRegister(func(db migrations.DB) (err error) {
		_, err = db.Exec(mg)
		return
	})
}
