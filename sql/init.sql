-- Create a trigger function to update the updated column
CREATE OR REPLACE FUNCTION update_updated()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- Create a macro to apply the trigger to a table
CREATE OR REPLACE FUNCTION apply_update_trigger(table_name TEXT)
RETURNS VOID AS $$
BEGIN
 IF NOT EXISTS (
    SELECT 1
    FROM information_schema.triggers
    WHERE trigger_schema = 'public'
      AND trigger_name = format('trigger_update_updated_%I', table_name)
  ) THEN
    EXECUTE format('
        CREATE TRIGGER trigger_update_updated_%I
        BEFORE UPDATE ON %I
        FOR EACH ROW
        EXECUTE FUNCTION update_updated()
    ', table_name, table_name);
  END IF;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS admins (
    id TEXT PRIMARY KEY,
    password TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS patterns (
    id TEXT PRIMARY KEY,
    query TEXT NOT NULL,
    search TEXT NOT NULL,
    dlpath TEXT NOT NULL,
    period TEXT NOT NULL,
    dayind TEXT NOT NULL,
    firetime TIMESTAMPTZ NOT NULL,
    created TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated TIMESTAMPTZ NOT NULL DEFAULT NOW(),
);

SELECT apply_update_trigger('patterns');
