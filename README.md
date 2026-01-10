# zavatar

**zavatar** is a lightweight avatar service that generates and serves user avatars based on user identity.  

---

## Modes

zavatar operates in two modes, determined by the presence of the `t` parameter.

### Service Mode (default)
- `t` not present
- Used for normal production traffic

### Preview Mode
- `t` present
- Intended for temporary or preview usage

---

## Avatar Types

The avatar type is selected by the numeric value of `t`.

| t | Type        | Description |
|---|-------------|------------|
| 1 | `identicon` | Pattern-based avatar derived from `user_id` |
| 2 | `letter`    | Text avatar derived from `username` |
| 3 | `gravatar` | External Gravatar integration |

---

## Sizes

Supported sizes:

```text
16, 32, 64, 128, 320
```

- Other sizes are normalized
- Avatars are generated at the target size

---

## API

```http
GET /u/{uid}?s={size}&t={type}
```

| Parameter | Required | Description |
|----------|----------|-------------|
| `uid` | yes | User ID |
| `s` | no | Size |
| `t` | no | Avatar type (numeric, enables preview mode) |

---

## Behavior

- Avatars are generated only for existing users
- Identical concurrent requests are deduplicated
- Generated or fetched avatars are served directly

---

## Configuration

### Server

```env
ADDR=:8080
BASE_URL=avatars.example.com
INTERNAL_KEY=xxxx
SITE_SALT=example.com
CF_ZONE_ID=xxxx
CF_API_TOKEN=xxxx
```

### Storage

```env
STORAGE_DRIVER=filesystem   # default

STORAGE_DRIVER=r2
R2_ACCOUNT_ID=xxxx
R2_BUCKET=zavatar
R2_ACCESS_KEY=xxxx
R2_SECRET_KEY=xxxx
R2_PREFIX=v1/
```

### Database

```env
DB_DRIVER=fake              # default

DB_DRIVER=mysql
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_USERNAME=root
MYSQL_PASSWORD=secret
MYSQL_DATABASE=pdb
MYSQL_USER_DATABASE=udb
```

---

## Examples

```text
# service mode
/u/1?s=32

# preview mode
/u/1?s=64&t=1
/u/1?s=64&t=2
/u/3?s=128&t=3
```
