CREATE TABLE IF NOT EXISTS books (
    id bigint generated always as identity primary key,
    title text not null,
    author text not null,
    year int not null,
    genre text not null,
    created_at TIMESTAMPTZ not null default now(),
    updated_at TIMESTAMPTZ not null default now()
)
