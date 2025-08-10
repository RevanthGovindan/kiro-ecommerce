-- Fix database schema issues
-- This script handles the foreign key constraint issue with user_id columns

-- Connect to the database first
\c ecommerce;

-- Drop foreign key constraints temporarily
ALTER TABLE addresses DROP CONSTRAINT IF EXISTS fk_users_addresses;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS fk_users_orders;

-- Drop existing tables if they exist (this will remove all data)
-- WARNING: This will delete all existing data
DROP TABLE IF EXISTS order_items CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS addresses CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- The application will recreate the tables with the correct schema
-- when it runs the GORM migrations