# zavatar

`zavatar` is a small Go HTTP service that generates and serves user avatars.

It supports multiple avatar “types” (e.g. letter, identicon, gravatar redirect) and can optionally store generated assets in a storage backend (local filesystem or Cloudflare R2). User profile data (which decides what avatar to use) can come from a DB backend (MySQL/MariaDB) or a fake/in-memory implementation for local dev/testing.

---

## Features

- **Letter avatar** (SVG)
- **Identicon avatar** (PNG)
- **Gravatar redirect** (HTTP redirect)
- **Storage backends**
  - local filesystem (`./data`)
  - Cloudflare R2 (S3-compatible)
- **DB backends**
  - MySQL/MariaDB
  - fake (for local/dev)
