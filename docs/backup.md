# Backup

One SQLite file: `DATA_DIR/pocket-horoscope.db` (WAL mode). Take a consistent copy with `pocket-horoscope backup` (writes `DATA_DIR/backups/pocket-horoscope-<time>.db` with `VACUUM INTO`) or Settings → Backup & data → Download database (admin). Back up that copy, not the live file. `DATA_DIR/core.token*` is the Pocket Core registration.

Restore: stop the app, replace `pocket-horoscope.db` (remove `-wal`/`-shm` files), start.

Every person can also export their own data as JSON (`/api/v1/export.json`: people, check-ins and kept readings; re-importable, importing twice adds nothing) or their people's birth data as CSV (`/api/v1/export.csv`, the columns the importer reads).
