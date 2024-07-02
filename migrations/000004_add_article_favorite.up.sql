CREATE TABLE IF NOT EXISTS article_favorite(
    id UUID CONSTRAINT article_favorite_pk PRIMARY KEY DEFAULT gen_random_uuid (),
    user_id UUID CONSTRAINT article_favorite_user_id_fk REFERENCES users (id) ON DELETE CASCADE,
    article_id UUID CONSTRAINT article_favorite_article_id_fk REFERENCES article (id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE CONSTRAINT article_favorite_created_at_df DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT article_favorite_user_id_article_id_uq UNIQUE (user_id, article_id)
)

     
     
     
     