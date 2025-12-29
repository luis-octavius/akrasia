CREATE TABLE todos (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT, 
  created_at DATE NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  concluded BOOLEAN NOT NULL 
);

