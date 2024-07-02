CREATE TABLE IF NOT EXISTS article_comment(
    id UUID CONSTRAINT article_comment_pk PRIMARY KEY DEFAULT gen_random_uuid (),
    author_id UUID CONSTRAINT article_comment_author_id_fk REFERENCES users (id) ON DELETE CASCADE,
    article_id UUID CONSTRAINT article_comment_article_id_fk REFERENCES article (id) ON DELETE CASCADE,
    body TEXT CONSTRAINT article_comment_body_nn NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE CONSTRAINT article_comment_created_at_df DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE CONSTRAINT article_comment_updated_at_df DEFAULT CURRENT_TIMESTAMP
);