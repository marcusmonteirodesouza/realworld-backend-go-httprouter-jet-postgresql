CREATE TABLE IF NOT EXISTS article (
    id UUID CONSTRAINT article_pk PRIMARY KEY DEFAULT gen_random_uuid (),
    author_id UUID CONSTRAINT article_author_id_fk REFERENCES users (id) ON DELETE CASCADE,
    slug TEXT CONSTRAINT article_slug_uk UNIQUE NOT NULL,
    title TEXT CONSTRAINT article_description_nn NOT NULL,
    description TEXT CONSTRAINT article_description_nn NOT NULL,
    body TEXT CONSTRAINT article_body_nn NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE CONSTRAINT article_created_at_df DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE CONSTRAINT article_updated_at_df DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS article_tag (
    id UUID CONSTRAINT article_tag_pk PRIMARY KEY DEFAULT gen_random_uuid (),
    name TEXT CONSTRAINT article_tag_name_uk UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE CONSTRAINT article_created_at_df DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS article_article_tag (
    id UUID CONSTRAINT article_article_tag_pk PRIMARY KEY DEFAULT gen_random_uuid (),
    article_id UUID CONSTRAINT article_article_tag_article_id REFERENCES article (id) ON DELETE CASCADE,
    article_tag_id UUID CONSTRAINT article_article_tag_article_tag_id REFERENCES article_tag (id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE CONSTRAINT article_created_at_df DEFAULT CURRENT_TIMESTAMP
);