ALTER TABLE users 
RENAME COLUMN confirmation_token TO token;

ALTER TABLE users 
DROP COLUMN confirmed_account;
