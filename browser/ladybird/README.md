# Ladybird

[Ladybird](https://ladybird.org/) is an independent web browser with its own engine.

## Cookie storage

Ladybird stores cookies in a SQLite database at:
- **Linux:** `~/.local/share/Ladybird/Ladybird.db` (or `$XDG_DATA_HOME/Ladybird/Ladybird.db`)
- **macOS:** `~/Library/Application Support/Ladybird/Ladybird.db`

## SQL schema of Ladybird.db, table Cookies

```sql
CREATE TABLE IF NOT EXISTS Cookies (
    name TEXT,
    value TEXT,
    same_site INTEGER CHECK (same_site >= 0 AND same_site <= 3),
    creation_time INTEGER,
    last_access_time INTEGER,
    expiry_time INTEGER,
    domain TEXT,
    path TEXT,
    secure BOOLEAN,
    http_only BOOLEAN,
    host_only BOOLEAN,
    persistent BOOLEAN,
    PRIMARY KEY(name, domain, path)
);
```

Timestamps are stored as Unix epoch **milliseconds**.
