BEGIN TRANSACTION;
DO $$
DECLARE
    last_version integer;
    new_version integer;
BEGIN
    new_version = 1;
    last_version = (SELECT version FROM db_version ORDER BY version ASC LIMIT 1);

    IF new_version != last_version+1 THEN
        RAISE EXCEPTION 'Wrong DB version';
    END IF;

    --
    -- INSERT YOUR MIGRATION CODE HERE
    --

    INSERT INTO db_version (version) VALUES (new_version);
END
$$ LANGUAGE plpgsql;
COMMIT;
