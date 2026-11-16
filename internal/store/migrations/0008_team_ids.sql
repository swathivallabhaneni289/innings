-- Fans can back more than one team (a country and an IPL side, say). team_ids is a JSON array of team ids, main team
-- first; favourite_team_id stays as the main team, so everything that shows one team keeps working. Additive: every
-- existing fan starts out backing just their favourite team. (|| rather than json_array keeps this portable.)

ALTER TABLE users ADD COLUMN team_ids TEXT NOT NULL DEFAULT '[]';
UPDATE users SET team_ids = '["' || favourite_team_id || '"]' WHERE favourite_team_id <> '';
