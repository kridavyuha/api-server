For Local Dev setup

1. After creating tables from schema.sql Run these

```
-- CSK Players
INSERT INTO players (player_id, player_name, team) VALUES
(0, 'AM Rahane', 'CSK'),
(1, 'DP Conway', 'CSK'),
(2, 'RD Gaikwad', 'CSK'),
(3, 'MS Dhoni', 'CSK'),
(4, 'RA Jadeja', 'CSK'),
(5, 'MM Ali', 'CSK'),
(6, 'S Dube', 'CSK'),
(7, 'TU Deshpande', 'CSK'),
(8, 'Akash Singh', 'CSK'),
(9, 'M Pathirana', 'CSK'),
(10, 'M Theekshana', 'CSK');
```



```
-- KKR Players
INSERT INTO players (player_id, player_name, team) VALUES
(11, 'AD Russell', 'KKR'),
(12, 'N Jagadeesan', 'KKR'),
(13, 'RK Singh', 'KKR'),
(14, 'JJ Roy', 'KKR'),
(15, 'CV Varun', 'KKR'),
(16, 'D Wiese', 'KKR'),
(17, 'K Khejroliya', 'KKR'),
(18, 'N Rana', 'KKR'),
(19, 'SP Narine', 'KKR'),
(20, 'Suyash Sharma', 'KKR'),
(21, 'VR Iyer', 'KKR'),
(22, 'UT Yadav', 'KKR');
```

```
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
```