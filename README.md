# zavatar

**zavatar** is a lightweight avatar service that resolves, generates, and serves user avatars based on user identity.

It exposes a single, stable avatar endpoint and internally selects the optimal strategy per avatar type
(e.g. direct serving, object storage redirection, or external redirection).

---

## Avatar Types

The avatar type is selected by `t` when provided. When `t` is omitted, the service uses the user's profile type from the API.

| t | Type        | Description |
|---|-------------|------------|
| 1 | `identicon` | Pattern-based avatar derived from `user_id` |
| 2 | `initials`   | Initials avatar derived from `user_name` |
| 3 | `gravatar` | External Gravatar avatar (redirected) |

- `t` is optional; omit it for the default profile type
- User-related data such as `user_name` is resolved internally from the API

---

## Sizes

Supported sizes:

```
16, 32, 64, 128, 320
```

- Other sizes are normalized to the next supported value (cap at 320)
- Avatars are generated at the final target size

---

## API

```
GET /u/{uid}?s={size}&t={type}
```

| Parameter | Required | Description |
|----------|----------|-------------|
| `uid` | yes | User ID |
| `s` | no | Avatar size |
| `t` | no | Avatar type (`1`, `2`, or `3`) |

---

## Behavior

### Common
- Avatars are served **only for existing users**
- Requests for non-existent users return `404`
- Identical concurrent requests are deduplicated internally

### t = 1 (identicon)
- Generated deterministically from `user_id`
- Generated avatars are stored
- Behavior depends on storage backend:
- `local`: avatar image is served directly by zavatar
  - `r2`: request is redirected to the R2 custom domain
- Long-term immutable caching is applied at the CDN level (when redirected)

### t = 2 (initials)
- Generated from the user's `user_name`
- User names are assumed to be stable
- Generated avatars are stored
- Behavior depends on storage backend:
- `local`: avatar image is served directly by zavatar
  - `r2`: request is redirected to the R2 custom domain
- Long-term immutable caching is applied at the CDN level (when redirected)

### t = 3 (gravatar)
- The user's `ghash` is resolved internally
- If `ghash` exists:
  - The request is redirected to the Gravatar URL
- If `ghash` does **not** exist:
  - The request automatically falls back to `t = 1` (identicon)
- No avatar data is stored by zavatar for this type

---

## Configuration

### Server

```
ADDR=:8080
SITE_SALT=example.com
```

### Storage

```
# Serve avatar images from local filesystem (default)
STORAGE_DRIVER=local

# Store avatar images in R2 and redirect requests to the custom domain
STORAGE_DRIVER=r2
R2_BUCKET=bucket1
R2_ACCOUNT_ID=xxxx
R2_ACCESS_KEY=xxxx
R2_SECRET_KEY=xxxx
R2_DIRECTORY=v1
R2_PUBLIC_BASE=https://avatars-cdn.example.com
```

### API

```
API_MODE=fake              # default

API_MODE=remote
API_ENDPOINT=https://api.example.com
API_SECRET_KEY=xxxx
```

---

## Examples

```
# identicon
/u/1?s=32&t=1

# initials
/u/1?s=64&t=2

# default type from API (t omitted)
/u/3?s=128

# gravatar (redirect or fallback)
/u/3?s=128&t=3
```
