CREATE TABLE IF NOT EXISTS players (
    player_id SERIAL PRIMARY KEY,
    player_name VARCHAR(255),
    team VARCHAR(255),
    year VARCHAR(4),
    league VARCHAR(255),
    profile_pic VARCHAR(255)
);

-- Common path prefix for assets
-- Base path: src/assets/
-- Team paths: CSK/, KKR/

-- CSK Players
INSERT INTO players (player_id, player_name, team) VALUES
('0', 'AM Rahane', 'CSK'),
('1', 'DP Conway', 'CSK'),
('2', 'RD Gaikwad', 'CSK'),
('3', 'MS Dhoni', 'CSK'),
('4', 'RA Jadeja', 'CSK'),
('5', 'MM Ali', 'CSK'),
('6', 'S Dube', 'CSK'),
('7', 'TU Deshpande', 'CSK'),
('8', 'Akash Singh', 'CSK'),
('9', 'M Pathirana', 'CSK'),
('10', 'M Theekshana', 'CSK');

-- KKR Players
INSERT INTO players (player_id, player_name, team) VALUES
('11', 'AD Russell', 'KKR'),
('12', 'N Jagadeesan', 'KKR'),
('13', 'RK Singh', 'KKR'),
('14', 'JJ Roy', 'KKR'),
('15', 'CV Varun', 'KKR'),
('16', 'D Wiese', 'KKR'),
('17', 'K Khejroliya', 'KKR'),
('18', 'N Rana', 'KKR'),
('19', 'SP Narine', 'KKR'),
('20', 'Suyash Sharma', 'KKR'),
('21', 'VR Iyer', 'KKR'),
('22', 'UT Yadav', 'KKR');

-- base_price table
CREATE TABLE IF NOT EXISTS base_price (
    player_id INT PRIMARY KEY,
    base_price INT
);

INSERT INTO base_price (player_id, base_price) VALUES
('0', 1500),
('1', 2000),
('2', 1800),
('3', 2200),
('4', 1700),
('5', 1600),
('6', 2100),
('7', 1900),
('8', 1400),
('9', 2300),
('10', 2500),
('11', 1200),
('12', 1300),
('13', 1100),
('14', 2400),
('15', 2000),
('16', 1500),
('17', 1800),
('18', 1700),
('19', 1600),
('20', 2100),
('21', 1900),
('22', 1400);




-- CSK Points
INSERT INTO points_1359507 (player_id, base_price, last_change) VALUES
(1, 1500, 'neu'),
(2, 2000, 'neu'),
(3, 2500, 'neu'),
(4, 3000, 'neu'),
(5, 3500, 'neu'),
(6, 1800, 'neu'),
(7, 2200, 'neu'),
(8, 2700, 'neu'),
(9, 3200, 'neu'),
(10, 3700, 'neu'),
(11, 1400, 'neu');

-- KKR Points
INSERT INTO points_1359507 (player_id, base_price, last_change) VALUES
(12, 1500, 'neu'),
(13, 2000, 'neu'),
(14, 2500, 'neu'),
(15, 3000, 'neu'),
(16, 3500, 'neu'),
(17, 1800, 'neu'),
(18, 2200, 'neu'),
(19, 2700, 'neu'),
(20, 3200, 'neu'),
(21, 3700, 'neu'),
(22, 1400, 'neu'),
(23, 1600, 'neu');

-- Drop points table
DROP TABLE IF EXISTS points_1359507;

UPDATE points_1359507
SET cur_price = base_price;





SELECT p.playerid, p.playername, p.team, p.profile_pic, pl.cur_price, pl.last_change
	FROM players p
	JOIN points_1359507 pl ON p.playerid = pl.player_id


	CREATE TABLE leagues (
		league_id SERIAL PRIMARY KEY,
		league_name VARCHAR(255)
	);
