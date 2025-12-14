-- Drop tracking tables to fix constraint issue
DROP TABLE IF EXISTS email_click_events CASCADE;
DROP TABLE IF EXISTS email_open_events CASCADE;
DROP TABLE IF EXISTS email_trackings CASCADE;
