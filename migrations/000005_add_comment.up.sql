CREATE TABLE IF NOT EXISTS comment(
    id UUID CONSTRAINT comment_pk PRIMARY KEY DEFAULT gen_random_uuid (),
    author_id UUID CONSTRAINT comment_author_id_fk REFERENCES users (id) ON DELETE CASCADE,
    article_id UUID CONSTRAINT comment_article_id_fk REFERENCES article (id) ON DELETE CASCADE,
    body TEXT CONSTRAINT comment_body_nn NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE CONSTRAINT comment_created_at_df DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE CONSTRAINT comment_updated_at_df DEFAULT CURRENT_TIMESTAMP
);