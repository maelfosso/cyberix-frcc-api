ALTER TABLE users
RENAME COLUMN token TO confirmation_token;

ALTER TABLE users
ADD COLUMN confirmed_account BOOLEAN DEFAULT FALSE;
