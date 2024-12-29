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
INSERT INTO players (player_name, team, year, league, profile_pic) VALUES
('AM Rahane', 'CSK', 2023, 'IPL', 'CSK/AM_Rahane.png'),
('DP Conway', 'CSK', 2023, 'IPL', 'CSK/DP_Conway.png'),
('RD Gaikwad', 'CSK', 2023, 'IPL', 'CSK/RD_Gaikwad.png'),
('MS Dhoni', 'CSK', 2023, 'IPL', 'CSK/MS_Dhoni.png'),
('RA Jadeja', 'CSK', 2023, 'IPL', 'CSK/RA_Jadeja.png'),
('MM Ali', 'CSK', 2023, 'IPL', 'CSK/MM_Ali.png'),
('S Dube', 'CSK', 2023, 'IPL', 'CSK/S_Dube.png'),
('TU Deshpande', 'CSK', 2023, 'IPL', 'CSK/TU_Deshpande.png'),
('Akash Singh', 'CSK', 2023, 'IPL', 'CSK/Akash_Singh.png'),
('M Pathirana', 'CSK', 2023, 'IPL', 'CSK/M_Pathirana.png'),
('M Theekshana', 'CSK', 2023, 'IPL', 'CSK/M_Theekshana.png');

-- KKR Players
INSERT INTO players (player_name, team, year, league, profile_pic) VALUES
('AD Russell', 'KKR', 2023, 'IPL', 'KKR/AD_Russell.png'),
('N Jagadeesan', 'KKR', 2023, 'IPL', 'KKR/N_Jagadeesan.png'),
('RK Singh', 'KKR', 2023, 'IPL', 'KKR/RK_Singh.png'),
('JJ Roy', 'KKR', 2023, 'IPL', 'KKR/JJ_Roy.png'),
('CV Varun', 'KKR', 2023, 'IPL', 'KKR/CV_Varun.png'),
('D Wiese', 'KKR', 2023, 'IPL', 'KKR/D_Wiese.png'),
('K Khejroliya', 'KKR', 2023, 'IPL', 'KKR/K_Khejroliya.png'),
('N Rana', 'KKR', 2023, 'IPL', 'KKR/N_Rana.png'),
('SP Narine', 'KKR', 2023, 'IPL', 'KKR/SP_Narine.png'),
('Suyash Sharma', 'KKR', 2023, 'IPL', 'KKR/Suyash_Sharma.png'),
('VR Iyer', 'KKR', 2023, 'IPL', 'KKR/VR_Iyer.png'),
('UT Yadav', 'KKR', 2023, 'IPL', 'KKR/UT_Yadav.png');




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
