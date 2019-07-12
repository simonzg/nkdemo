
INSERT INTO users (id, seq, handle, lang,custom_id, created_at, updated_at) values ('\x706466494a677a67536a36634c514e484f7456434267', 1, 'DEdhMekgnp', 'en', 'test', 1515007683339, 1515007683339);

INSERT INTO assets (id, thumbnail_url, "name", url, type, category, version, metadata, created_at, updated_at, thumbnail_version) VALUES
	('\x31', 'http://52.53.222.194/file/xbot.png', 'xbot', 'http://52.53.222.194/file/xbot.dlc', 'avatar', ' ', 1, '\x7b7d', 1516908523, 1516908523, 0),
	('\x32', 'http://52.53.222.194/file/test-space.png', 'theme', 'http://52.53.222.194/file/test-space.dlc', 'theme', ' ', 1, '\x7b7d', 1516908523, 1516908523, 0);


INSERT INTO spaces (id, seq, creator_id, display_name, description, theme, thumbnail_url, lang, utc_offset_ms, metadata, state, count, created_at, updated_at, disabled_at, asset_id) VALUES
	('\x796b617935785035516375687033536e7163334d5177', 100002, '\x706466494a677a67536a36634c514e484f7456434267', 'test-space', 'Test Space Description', '', 'http://52.53.222.194/file/testSpace.png', 'en', 0, '\x7b7d', 0, 1, 1515007683339, 1515007683339, 0, '\x32');
