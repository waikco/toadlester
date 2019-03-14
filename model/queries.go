package model

//CreateTableQuery is sql query for creating fda_data table
const CreateTableQuery string = `CREATE TABLE IF NOT EXISTS tests (
id SERIAL PRIMARY KEY,
name TEXT NOT NULL,
data jsonb);`

//TestCreateTableQuery is sql query for creating fda_data table
const TestCreateTableQuery string = `CREATE TABLE IF NOT EXISTS tests(
id SERIAL PRIMARY KEY,
name TEXT NOT NULL,
data jsonb
);`

//HealthQuery is used for health checks to test database connection
const HealthQuery string = `SELECT 1`
