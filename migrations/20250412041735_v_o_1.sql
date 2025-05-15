-- +goose Up
-- +goose StatementBegin
insert into users (id,auth0_sub, username,  user_type,bot_type, user_meta, premium, is_active, is_deleted, created_on, updated_on, created_by, updated_by) values ('00000000-0000-0000-0000-000000000001',NULL,'admin', 'HUMAN',null, '{}', true, true, false, now(), now(), 'admin', 'admin');
insert into users (id, auth0_sub, username,  user_type,bot_type, user_meta, premium, is_active, is_deleted, created_on, updated_on, created_by, updated_by) values ('00000000-0000-0000-0000-000000000002', NULL,'Sec10', 'BOT','10 sec', '{}', true, true, false, now(), now(), 'admin', 'admin');
insert into users (id, auth0_sub, username,  user_type,bot_type, user_meta, premium, is_active, is_deleted, created_on, updated_on, created_by, updated_by) values ('00000000-0000-0000-0000-000000000003', NULL, 'Sec15', 'BOT','15 sec', '{}', true, true, false, now(), now(), 'admin', 'admin');
insert into users (id, auth0_sub, username,  user_type,bot_type, user_meta, premium, is_active, is_deleted, created_on, updated_on, created_by, updated_by) values ('00000000-0000-0000-0000-000000000004', NULL, 'Sec20', 'BOT','20 sec', '{}', true, true, false, now(), now(), 'admin', 'admin');


insert into users (id, auth0_sub, username,  user_type,bot_type, user_meta, premium, is_active, is_deleted, created_on, updated_on, created_by, updated_by) values ('00000000-0000-0000-0000-000000000005', NULL,'Sec30', 'BOT','30 sec', '{}', true, true, false, now(), now(), 'admin', 'admin');
insert into users (id, auth0_sub, username,  user_type,bot_type, user_meta, premium, is_active, is_deleted, created_on, updated_on, created_by, updated_by) values ('00000000-0000-0000-0000-000000000006', NULL,'Sec45', 'BOT','45 sec', '{}', true, true, false, now(), now(), 'admin', 'admin');
insert into users (id, auth0_sub, username,  user_type,bot_type, user_meta, premium, is_active, is_deleted, created_on, updated_on, created_by, updated_by) values ('00000000-0000-0000-0000-000000000007',NULL, 'Sec1', 'BOT','1 min', '{}', true, true, false, now(), now(), 'admin', 'admin');
insert into users (id, auth0_sub, username,  user_type,bot_type, user_meta, premium, is_active, is_deleted, created_on, updated_on, created_by, updated_by) values ('00000000-0000-0000-0000-000000000008',NULL, 'Sec2', 'BOT','2 min', '{}', true, true, false, now(), now(), 'admin', 'admin');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000001';
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000002';
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000003';
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000004';
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000005';
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000006';
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000007';
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000008';

-- +goose StatementEnd
