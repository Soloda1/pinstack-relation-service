CREATE TABLE followers (
   id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
   follower_id BIGINT NOT NULL,
   followee_id BIGINT NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

   CONSTRAINT unique_follower_followee UNIQUE (follower_id, followee_id)
);

CREATE INDEX idx_follower_id ON followers(follower_id);
CREATE INDEX idx_followee_id ON followers(followee_id);


CREATE TABLE outbox (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    aggregate_id BIGINT NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'pending', 'sent', 'error')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at TIMESTAMPTZ
);
CREATE INDEX idx_outbox_status ON outbox(status);