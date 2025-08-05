-- USERS
INSERT INTO USERS (user_email, user_name, user_hash, user_description, user_score, user_isadmin, user_isbot, user_ishidden, user_isdeactivated, user_mustchangename, user_mustchangepassword, user_invited)
VALUES
('cannanbot@example.com', 'CannanBot', 'hash1', 'Assistant (to the) manager.', 0, 1, 1, 1, 0, 0, 0, '2025-08-04 10:00:00'),
('lee@example.com', 'Lee', 'hash2', 'Site manager', 0, 1, 0, 1, 0, 1, 0, '2025-08-04 11:00:00'),
('eve@example.com', 'Eve', 'hash3', 'Test user', 7, 0, 0, 0, 0, 0, 0, '2025-08-04 12:00:00'),
('carol@example.com', 'Carol', 'hash4', 'Test user', 9, 0, 0, 0, 0, 0, 0, '2025-08-04 13:00:00'),
('dave@example.com', 'Dave', 'hash5', 'Test user', 3, 0, 0, 0, 0, 0, 0, '2025-08-04 14:00:00');

-- IPS
INSERT INTO IPS (user_id, ip_address, ip_first_seen, ip_days_seen, ip_note)
VALUES
(1, '127.0.0.1', '2025-08-04 09:00:00', 4, 'Server Bot'),
(2, '10.0.0.4', '2025-08-03 18:00:00', 3, 'Lee laptop'),
(3, '172.16.5.7', '2025-08-02 15:00:00', 1, 'Eve proxy'),
(4, '192.168.1.10', '2025-08-04 11:00:00', 5, 'Carol office'),
(5, '10.1.2.3', '2025-08-02 17:30:00', 2, 'Dave VPN');

-- CHALLENGES
INSERT INTO CHALLENGES (challenge_title, challenge_tags, challenge_description, challenge_points, challenge_flag, challenge_hidden)
VALUES
('Caesar Salad', 'crypto, beginner', 'Someone sent us this message, but it looks a bit scrambled. Can you help decipher it?\n\nWklv lv d whvw phvvdjh\n\nHint: Try shifting each letter backwards.', 1, 'flag{this_is_a_test_message}', 0),
('Bacon Bits', 'crypto, easy', 'We intercepted this binary-looking string. It doesn’t taste like bacon, but maybe it’s hiding something:\n\nAABBA AABAA ABBAB ABAAB AABBA AABAA\n\nHint: A=‘A’, B=‘B’. That’s all you need.', 2, 'flag{bacon_is_tasty}', 0),
('Subbed Out', 'crypto, substitution', 'A secret agent left behind this message. It''s not a simple shift — something fancier is at play:\n\nZOL SLHYU JVTWHSS\n\nHint: Maybe "LEARN" something from it.', 3, 'flag{you_cracked_the_code}', 0),
('RSA Rookie', 'crypto, rsa', 'We found this RSA-encrypted message. Lucky for you, the key is small...\n\nn = 2537\ne = 13\ncipher = 2000\n\nHint: Factor n, find d, decrypt the message.', 4, 'flag{rsa_is_fun}', 0),
('Obscured Vision', 'crypto, vigenere', 'The enemy is using a Vigenère cipher to hide communications. Here''s what we intercepted:\n\nEncrypted: ZICVTWQNGRZGVTWAVZHCQYGLMGJ\nKeyword: CRYPTO\n\nHint: Repeating key, repeating patterns.', 5, 'flag{vigenere_decrypted_message}', 1);

-- SOLVES
INSERT INTO SOLVES (user_id, challenge_id)
VALUES
(3, 1), (3, 2), (3, 4),
(4, 2), (4, 3), (4, 4),
(5, 1), (5, 2);

-- NOTICES
INSERT INTO NOTICES (user_id, notice_title, notice_content)
VALUES
(1, 'Welcome', 'Welcome to the platform!'),
(2, 'Maintenance', 'System will be down tonight.'),
(3, 'Flag Policy', 'Do not share flags.'),
(4, 'Writeups', 'Share your writeups!'),
(5, 'Leaderboard', 'Updated leaderboard posted.'),
(1, 'Hints', 'Some new hints are up.'),
(2, 'CTF Closing', 'The event ends tomorrow.'),
(3, 'Update', 'We pushed a minor bug fix.'),
(4, 'Rules Reminder', 'Stay ethical!'),
(5, 'Discord', 'Join us on Discord.');

-- NOTICES_QUEUE
INSERT INTO NOTICES_QUEUE (user_id, notice_queue_title, notice_queue_content, notice_queue_scheduled_datetime, notice_queue_crosspost_to_discord)
VALUES
(1, 'Upcoming CTF', 'New CTF launching soon!', '2025-08-05 10:00:00', 1),
(2, 'Downtime', 'Scheduled downtime.', '2025-08-06 00:00:00', 0),
(3, 'Flag Update', 'Fixed flag typos.', '2025-08-04 14:00:00', 1),
(4, 'Leaderboard Drop', 'Top 5 updated.', '2025-08-04 18:00:00', 1),
(5, 'Survey', 'Feedback survey open.', '2025-08-05 12:00:00', 0),
(1, 'Finals', 'Final challenge unlocks!', '2025-08-07 09:00:00', 1),
(2, 'Patch Notes', 'Backend fixes rolled out.', '2025-08-03 11:00:00', 0),
(3, 'Discord Poll', 'Vote for next theme.', '2025-08-08 16:00:00', 1),
(4, 'Live Stream', 'Tune in live.', '2025-08-09 19:00:00', 1),
(5, 'Bot Update', 'Oscar upgraded.', '2025-08-02 13:00:00', 0);

-- MESSAGES
INSERT INTO MESSAGES (user_id, message_content)
VALUES
(1, 'Good luck all!'),
(2, 'Cant believe I solved that.'),
(3, 'Oops, wrong flag.'),
(4, 'Anyone stuck on challenge 5?'),
(5, 'Shoutout to Alice for help.'),
(1, 'This XSS one is tough.'),
(2, 'Just pwned the binary!'),
(3, 'Thanks for the hints.'),
(4, 'Bot here. Nothing to see.'),
(5, 'Final boss is impossible.');

-- MESSAGES_QUEUE
INSERT INTO MESSAGES_QUEUE (user_id, message_queue_content, message_queue_scheduled_datetime, message_queue_crosspost_to_discord)
VALUES
(1, 'CTF starts tomorrow!', '2025-08-05 08:00:00', 1),
(2, 'Flag submission fixed.', '2025-08-06 09:00:00', 0),
(3, 'Bot update complete.', '2025-08-04 18:00:00', 1),
(4, 'Leaderboard changes live.', '2025-08-05 13:00:00', 0),
(5, 'Hint drop coming.', '2025-08-06 10:00:00', 1),
(1, 'Endgame unlocks soon.', '2025-08-07 20:00:00', 1),
(2, 'Congrats top players.', '2025-08-08 09:00:00', 0),
(3, 'Theme poll live.', '2025-08-08 15:00:00', 1),
(4, 'Discord AMA today.', '2025-08-09 12:00:00', 1),
(5, 'Thanks for playing!', '2025-08-10 18:00:00', 1);

-- CONFIG
INSERT INTO CONFIG (user_id, config_parameter, config_value)
VALUES
(2, 'site_title', 'MyCTF'),
(2, 'max_login_attempts', '5'),
(2, 'discord_webhook_url', 'https://discord.example.com/webhook'),
(2, 'flag_format', 'flag{.*}'),
(2, 'maintenance_mode', 'off'),
(2, 'theme', 'dark'),
(2, 'allow_registration', 'true'),
(2, 'default_score', '0'),
(2, 'hint_cost', '10'),
(2, 'final_challenge_id', '10');