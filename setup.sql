-- run this on your MSSQL server before starting the app
-- creates the database, table, and stored procedures

CREATE DATABASE weather_loc_service;
GO

USE weather_loc_service;
GO

CREATE TABLE query_history (
    id INT IDENTITY(1,1) PRIMARY KEY,
    query_type VARCHAR(20) NOT NULL,
    query_text VARCHAR(500) NOT NULL,
    latitude DECIMAL(10,6) NOT NULL,
    longitude DECIMAL(10,6) NOT NULL,
    response_summary VARCHAR(500),
    created_at DATETIME2 DEFAULT GETDATE()
);
GO

CREATE INDEX idx_history_date ON query_history(created_at);
CREATE INDEX idx_history_type ON query_history(query_type);
GO

-- stored procedure: insert a query record
CREATE OR ALTER PROCEDURE sp_InsertQuery
    @query_type VARCHAR(20),
    @query_text VARCHAR(500),
    @latitude DECIMAL(10,6),
    @longitude DECIMAL(10,6),
    @response_summary VARCHAR(500)
AS
BEGIN
    SET NOCOUNT ON;
    INSERT INTO query_history (query_type, query_text, latitude, longitude, response_summary)
    VALUES (@query_type, @query_text, @latitude, @longitude, @response_summary);

    SELECT SCOPE_IDENTITY() AS id;
END;
GO

-- stored procedure: get query history with optional type filter
CREATE OR ALTER PROCEDURE sp_GetQueryHistory
    @query_type VARCHAR(20) = NULL,
    @limit INT = 50
AS
BEGIN
    SET NOCOUNT ON;
    SELECT TOP(@limit) id, query_type, query_text, latitude, longitude, response_summary, created_at
    FROM query_history
    WHERE (@query_type IS NULL OR query_type = @query_type)
    ORDER BY created_at DESC;
END;
GO

-- stored procedure: get query stats grouped by type
CREATE OR ALTER PROCEDURE sp_GetQueryStats
AS
BEGIN
    SET NOCOUNT ON;
    SELECT query_type, COUNT(*) AS total_count
    FROM query_history
    GROUP BY query_type
    ORDER BY total_count DESC;
END;
GO
