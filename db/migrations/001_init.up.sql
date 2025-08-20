CREATE TABLE IF NOT EXISTS subscriptions (
                                             service_name TEXT NOT NULL,
                                             price INTEGER NOT NULL,
                                             user_id UUID NOT NULL,
                                             start_date TIMESTAMP NOT NULL,
                                             end_date TIMESTAMP,
                                             PRIMARY KEY (user_id, service_name)
);
