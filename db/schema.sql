-- So i am going with this architectural decision of creating new table for each league if it can be deleted after the contest.


-- Create players table
-- load all these players data even before tournament
-- so that we can assign base_prices as well in prior.
-- As we only open the league to users after toss, we just fetch the squad from the 
-- api and add them to players and base price and recreate the leagues (RARE case)...
CREATE TABLE IF NOT EXISTS players (
    player_id VARCHAR(6) PRIMARY KEY,
    player_name VARCHAR(255) NOT NULL,
    team VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


-- assign base price for all the players manually...
-- This has to be done even before the match.
Create Table base_price (
    player_id VARCHAR(6) PRIMARY KEY,
    base_price INT NOT NULL,
    FOREIGN KEY (player_id) REFERENCES players(player_id)
);


-- both the above tables are to be created before the tournment.



-- Create Leagues table  -> (league details)
-- Add Leagues here once a post request is made with the required team_ids, entry_fee, capacity, match_id.
-- users_registered is a comma separated user_ids...
CREATE TABLE leagues (
    league_id VARCHAR(100) PRIMARY KEY,
    match_id VARCHAR(50) NOT NULL,
    entry_fee INT NOT NULL,
    capacity INT NOT NULL DEFAULT 100,
    registered INT DEFAULT 0,
    users_registered TEXT DEFAULT '',
    league_status VARCHAR(15) DEFAULT 'upcoming' CHECK (league_status IN ('active', 'completed', 'upcoming')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- once this table is created we also create a points_{} table to track the cur_price.
-- so here we call squads api once to get the player id's belonging to those teams and get their respective base prices.


-- Create points table with league_id (Not to be created)
-- TODO: Add a foriegn key player_id to players table.

-- This table gets created automatically, don't create it
-- CREATE TABLE players_{league_name} (
--     player_id SERIAL PRIMARY KEY,
--     base_price INT,
--     cur_price INT,
--     last_change VARCHAR(3) CHECK (last_change IN ('pos', 'neg', 'neu')),
-- );



-- Create Users table 
CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    user_name VARCHAR(50) NOT NULL,
    mail_id VARCHAR(100) NOT NULL,
    profile_pic VARCHAR(100),
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert a new user into the users table
INSERT INTO users (user_name, mail_id, profile_pic, password) 
VALUES ('John Doe', 'john.doe@example.com', 'profile_pic_url', 'password');


-- Create Purse table to store the remaining purse for each user in each league.
CREATE TABLE purse (
    user_id INT NOT NULL,
    league_id VARCHAR(100) NOT NULL,
    remaining_purse INT DEFAULT 10000,
    PRIMARY KEY (user_id, league_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (league_id) REFERENCES leagues(league_id)
);

CREATE TABLE portfolio (
    user_id INT NOT NULL,
    league_id VARCHAR(100) NOT NULL,
    player_id VARCHAR(6) NOT NULL,
    shares INT DEFAULT 0,
    PRIMARY KEY (user_id, league_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (league_id) REFERENCES leagues(league_id),
    FOREIGN KEY (player_id) REFERENCES players(player_id)
);


-- Create a table to store the transactions of the users.
-- As this is a bit rarely acccessed table, we can keep it in the same table.
Create Table transactions (
    user_id INT NOT NULL,
    league_id VARCHAR(100) NOT NULL,
    player_id VARCHAR(6) NOT NULL,
    shares INT NOT NULL,
    price INT NOT NULL,
    transaction_type VARCHAR(10) CHECK (transaction_type IN ('buy', 'sell')),
    transaction_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (league_id) REFERENCES leagues(league_id),
    FOREIGN KEY (player_id) REFERENCES players(player_id)
);


