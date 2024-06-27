CREATE TABLE IF NOT EXISTS favorite(
    id UUID CONSTRAINT favorite_pk PRIMARY KEY DEFAULT gen_random_uuid (),
    user_id UUID CONSTRAINT favorite_user_id_fk REFERENCES users (id) ON DELETE CASCADE,
    article_id UUID CONSTRAINT favorite_article_id_fk REFERENCES article (id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE CONSTRAINT favorite_created_at_df DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT favorite_user_id_article_id_uq UNIQUE (user_id, article_id)
)

     
     
     
     