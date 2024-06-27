CREATE TABLE IF NOT EXISTS follow (
     id UUID CONSTRAINT follow_pk PRIMARY KEY DEFAULT gen_random_uuid (),
     follower_id UUID CONSTRAINT follow_follower_id_fk REFERENCES users (id) ON DELETE CASCADE,
     followed_id UUID CONSTRAINT follow_followed_id_fk REFERENCES users (id) ON DELETE CASCADE,
     created_at TIMESTAMP WITH TIME ZONE CONSTRAINT follow_created_at_df DEFAULT CURRENT_TIMESTAMP,
     CONSTRAINT follow_follower_id_followed_id_uq UNIQUE (follower_id, followed_id)
);