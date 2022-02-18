CREATE TABLE urls (
       id SERIAL PRIMARY KEY,
       token varchar(50) UNIQUE NOT NULL,
       url varchar(2048) NOT NULL
);