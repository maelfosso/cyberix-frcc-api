ALTER TABLE users
ADD current_otp TEXT,
ADD current_otp_validity_time TIMESTAMPZ DEFAULT NOW() + INTERVAL '2 minutes';
