# Firefox

## SQL schemes of file cookies.sqlite, table moz_cookies

extracted with sqlitebrowser

```sql
-- Firefox 91 ESR Linux, Firefox 98.0.2 Windows 7
CREATE TABLE moz_cookies (
    id INTEGER PRIMARY KEY,
    originAttributes TEXT NOT NULL DEFAULT '',
    name TEXT,
    value TEXT,
    host TEXT,
    path TEXT,
    expiry INTEGER,
    lastAccessed INTEGER,
    creationTime INTEGER,
    isSecure INTEGER,
    isHttpOnly INTEGER,
    inBrowserElement INTEGER DEFAULT 0,
    sameSite INTEGER DEFAULT 0,
    rawSameSite INTEGER DEFAULT 0,
    schemeMap INTEGER DEFAULT 0,
    CONSTRAINT moz_uniqueid UNIQUE (name, host, path, originAttributes)
)
```

```sql
-- Firefox 78 ESR Linux
CREATE TABLE moz_cookies(
    id INTEGER PRIMARY KEY,
    originAttributes TEXT NOT NULL DEFAULT '',
    name TEXT,
    value TEXT,
    host TEXT,
    path TEXT,
    expiry INTEGER,
    lastAccessed INTEGER,
    creationTime INTEGER,
    isSecure INTEGER,
    isHttpOnly INTEGER,
    inBrowserElement INTEGER DEFAULT 0,
    sameSite INTEGER DEFAULT 0,
    rawSameSite INTEGER DEFAULT 0,
    CONSTRAINT moz_uniqueid UNIQUE (name, host, path, originAttributes)
)
```

## Cookie expiry units: seconds vs milliseconds

SQLite `user_version` pragma:

- `user_version <= 15` → `moz_cookies.expiry` is **Unix seconds** → `time.Unix(expiry, 0)`
- `user_version >= 16` → `moz_cookies.expiry` is **Unix milliseconds** → `time.UnixMilli(expiry)`

Firefox 141 and earlier (schema ≤ 15) store seconds.
Firefox 142 introduced schema 16 and migrated existing values with
`UPDATE moz_cookies SET expiry = expiry * 1000` (Mozilla Bug 1972757).

- Firefox 141, `user_version` 15, no `expiry` migration:
  <https://hg.mozilla.org/releases/mozilla-release/raw-file/FIREFOX_141_0_RELEASE/netwerk/cookie/CookiePersistentStorage.cpp>
- Firefox 142, `user_version` 16, `expiry = expiry * 1000`:
  <https://hg.mozilla.org/releases/mozilla-release/raw-file/FIREFOX_142_0_RELEASE/netwerk/cookie/CookiePersistentStorage.cpp>
- Firefox 141, `Cookie::Expiry()` documented as seconds:
  <https://hg.mozilla.org/releases/mozilla-release/raw-file/FIREFOX_141_0_RELEASE/netwerk/cookie/Cookie.h>
- Firefox 142, `Cookie::Expiry()` documented as milliseconds:
  <https://hg.mozilla.org/releases/mozilla-release/raw-file/FIREFOX_142_0_RELEASE/netwerk/cookie/Cookie.h>

## Later schema versions

Firefox 146 bumped `user_version` to 17, adding an `updateTime` column:
<https://hg.mozilla.org/releases/mozilla-release/raw-file/FIREFOX_146_0_RELEASE/netwerk/cookie/CookiePersistentStorage.cpp>
