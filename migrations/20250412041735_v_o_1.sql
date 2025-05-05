-- +goose Up
-- +goose StatementBegin
insert into users (id, username, refresh_token, user_type,bot_type, user_meta, premium, is_active, is_deleted, created_on, updated_on, created_by, updated_by) values ('00000000-0000-0000-0000-000000000001', 'admin', 'admin', 'HUMAN',null, '{}', true, true, false, now(), now(), 'admin', 'admin');
insert into users (id, username, refresh_token, user_type,bot_type, user_meta, premium, is_active, is_deleted, created_on, updated_on, created_by, updated_by) values ('00000000-0000-0000-0000-000000000002', 'bot-1', 'Sec10', 'BOT','10 sec', '{}', true, true, false, now(), now(), 'admin', 'admin');
insert into users (id, username, refresh_token, user_type,bot_type, user_meta, premium, is_active, is_deleted, created_on, updated_on, created_by, updated_by) values ('00000000-0000-0000-0000-000000000003', 'bot-2', 'Sec15', 'BOT','15 sec', '{}', true, true, false, now(), now(), 'admin', 'admin');
insert into users (id, username, refresh_token, user_type,bot_type, user_meta, premium, is_active, is_deleted, created_on, updated_on, created_by, updated_by) values ('00000000-0000-0000-0000-000000000004', 'bot-4', 'Sec20', 'BOT','20 sec', '{}', true, true, false, now(), now(), 'admin', 'admin');


insert into users (id, username, refresh_token, user_type,bot_type, user_meta, premium, is_active, is_deleted, created_on, updated_on, created_by, updated_by) values ('00000000-0000-0000-0000-000000000005', 'bot-5', 'Sec30', 'BOT','30 sec', '{}', true, true, false, now(), now(), 'admin', 'admin');
insert into users (id, username, refresh_token, user_type,bot_type, user_meta, premium, is_active, is_deleted, created_on, updated_on, created_by, updated_by) values ('00000000-0000-0000-0000-000000000006', 'bot-6', 'Sec45', 'BOT','45 sec', '{}', true, true, false, now(), now(), 'admin', 'admin');
insert into users (id, username, refresh_token, user_type,bot_type, user_meta, premium, is_active, is_deleted, created_on, updated_on, created_by, updated_by) values ('00000000-0000-0000-0000-000000000007', 'bot-7', 'Sec1', 'BOT','1 min', '{}', true, true, false, now(), now(), 'admin', 'admin');
insert into users (id, username, refresh_token, user_type,bot_type, user_meta, premium, is_active, is_deleted, created_on, updated_on, created_by, updated_by) values ('00000000-0000-0000-0000-000000000008', 'bot-8', 'Sec2', 'BOT','2 min', '{}', true, true, false, now(), now(), 'admin', 'admin');


-- +goose StatementEnd

-- +goose Down
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000005';
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000006';
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000007';
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000008';

ALTER TABLE room_member DROP COLUMN room_id;
ALTER TABLE leaderboard DROP COLUMN is_deleted;
-- +goose StatementBegin
-- +goose StatementEnd
