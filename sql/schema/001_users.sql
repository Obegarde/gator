-- +goose Up
CREATE TABLE  users (
	id UUID PRIMARY KEY,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	name VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE feeds(
	id UUID PRIMARY KEY,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	user_id UUID NOT NULL REFERENCES  users (id) ON DELETE CASCADE,
	name VARCHAR(255) NOT NULL,
	url VARCHAR(255) UNIQUE NOT NULL

);

CREATE TABLE feed_follows(
	id UUID PRIMARY KEY,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	user_id UUID,
	feed_id UUID,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
	UNIQUE (user_id, feed_id)
);
-- +goose Down
DROP TABLE feed_follows;
DROP TABLE feeds;
DROP TABLE users;
