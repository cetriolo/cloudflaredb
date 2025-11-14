-- Cloudflare D1 Database Cleanup Script
-- This script drops all tables to allow fresh migrations to run
--
-- Usage:
--   npx wrangler d1 execute cloudflaredb --remote --file=scripts/cleanup-d1.sql
--   OR
--   npx wrangler d1 execute cloudflaredb --local --file=scripts/cleanup-d1.sql
--
-- WARNING: This will delete all data in your database!

-- Drop tables in reverse dependency order to avoid foreign key constraint issues
DROP TABLE IF EXISTS user_rooms;
DROP TABLE IF EXISTS rooms;
DROP TABLE IF EXISTS room_types;
DROP TABLE IF EXISTS users;

-- Drop indexes (if they weren't dropped with tables)
DROP INDEX IF EXISTS idx_users_external_id;
DROP INDEX IF EXISTS idx_room_types_size;
DROP INDEX IF EXISTS idx_room_types_style;
DROP INDEX IF EXISTS idx_rooms_room_type_id;
DROP INDEX IF EXISTS idx_user_rooms_user_id;
DROP INDEX IF EXISTS idx_user_rooms_room_id;

-- Verify cleanup
SELECT 'Database cleanup complete. All tables dropped.' AS status;
