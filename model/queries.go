package model

//FoodDataCreateTableQuery is sql query for creating fda_data table
const SqlFoodDataCreateTableQuery string = `CREATE TABLE IF NOT EXISTS fda_data
(
id UUID PRIMARY KEY,
name TEXT NOT NULL,
data jsonb,
)`

//TestFoodDataCreateTableQuery is sql query for creating fda_data table
const SqlTestFoodDataCreateTableQuery string = `use testspace; create table govdata 
(
id UUID PRIMARY KEY,
name TEXT NOT NULL,
data jsonb,
);`

//HealthQuery is used for health checks to test database connection
const HealthQuery string = `SELECT 1`
