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
