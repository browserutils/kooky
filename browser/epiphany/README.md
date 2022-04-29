### SQL schemes of file cookies.sqlite, table moz_cookies

extracted with sqlitebrowser

```sql
-- epiphany (Gnome Web) 3.38.2 Linux
CREATE TABLE moz_cookies (
    id INTEGER PRIMARY KEY,
    name TEXT,
    value TEXT,
    host TEXT,
    path TEXT,
    expiry INTEGER,
    lastAccessed INTEGER,
    isSecure INTEGER,
    isHttpOnly INTEGER,
    sameSite INTEGER -- new
)
```

```sql
-- epiphany (Gnome Web) 3.32 Linux
CREATE TABLE moz_cookies (
    id INTEGER PRIMARY KEY,
    name TEXT,
    value TEXT
    host TEXT,
    path TEXT,
    expiry INTEGER,
    lastAccessed INTEGER,
    isSecure INTEGER,
    isHttpOnly INTEGER
)
```
