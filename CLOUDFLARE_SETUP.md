# Cloudflare Setup Guide – occ.io.vn → KidTyping VN

## Overview

| Item | Value |
|------|-------|
| Domain | `occ.io.vn` (+ `www.occ.io.vn`) |
| VPS IP | `51.79.138.246` |
| App port on VPS | `11100` |
| SSL on VPS | Cloudflare Origin Certificate (valid to 2041, covers `*.occ.io.vn` + `occ.io.vn`) |

---

## 1. DNS Records

In the Cloudflare Dashboard → **occ.io.vn** → **DNS → Records**, ensure these two records exist:

| Type | Name | Content | Proxy status | TTL |
|------|------|---------|--------------|-----|
| A | `occ.io.vn` (or `@`) | `51.79.138.246` | **Proxied** (orange cloud ☁️) | Auto |
| A | `www` | `51.79.138.246` | **Proxied** (orange cloud ☁️) | Auto |

> **Why orange cloud?** The VPS uses a Cloudflare Origin Certificate which is **only trusted by Cloudflare** (not browsers). You must keep the proxy enabled so Cloudflare terminates TLS for visitors and forwards via its own trusted certificate.

---

## 2. SSL / TLS Mode

Go to **SSL/TLS → Overview** and set the mode to:

> ✅ **Full (strict)**

| Mode | Meaning |
|------|---------|
| Off | HTTP only – never use |
| Flexible | Cloudflare → VPS over plain HTTP (insecure) |
| Full | Cloudflare → VPS over HTTPS but certificate not validated |
| **Full (strict)** | Cloudflare → VPS over HTTPS, certificate fully validated ← **use this** |

The VPS nginx already has the Cloudflare Origin Certificate installed at:
- `/data/nginx-proxy/ssl/occ.io.vn.pem.cloudflare.bak`
- `/data/nginx-proxy/ssl/occ.io.vn.key.cloudflare.bak`

These are referenced in `/data/nginx-proxy/conf/typing-for-kids.conf`.

---

## 3. Minimum Recommended Security Settings

### SSL/TLS → Edge Certificates
- **Always Use HTTPS**: ON
- **Minimum TLS Version**: TLS 1.2
- **Opportunistic Encryption**: ON
- **TLS 1.3**: ON
- **Automatic HTTPS Rewrites**: ON

### Security → Settings  
- **Browser Integrity Check**: ON
- **Security Level**: Medium

---

## 4. Page Rules (optional but recommended)

Add a Page Rule to force HTTPS for any leftover HTTP requests if "Always Use HTTPS" doesn't catch them:

| URL pattern | Setting | Value |
|-------------|---------|-------|
| `http://occ.io.vn/*` | Always Use HTTPS | — |
| `http://www.occ.io.vn/*` | Always Use HTTPS | — |

---

## 5. Verify Everything Works

After applying Cloudflare settings, test from your own browser or command line:

```bash
# Should return HTTP 200 and the KidTyping page title
curl -sI https://occ.io.vn | head -5
curl -s https://occ.io.vn | grep '<title>'
```

Expected output:
```
<title>KidTyping VN – Luyện Gõ Tiếng Việt</title>
```

---

## 6. Troubleshooting

### Error 521 – Web server is down
- Check the app is running: `ssh vps "systemctl status kidtyping"`
- Restart if needed: `ssh vps "systemctl restart kidtyping"`

### Error 525 – SSL handshake failed  
- SSL/TLS mode is set to **Full (strict)** but cert mismatch.  
- Verify the correct cert files are referenced in `/data/nginx-proxy/conf/typing-for-kids.conf`.

### Error 526 – Invalid SSL certificate
- Cloudflare cannot validate the origin cert.
- Make sure you use the **Cloudflare Origin Certificate** (the `.cloudflare.bak` files), not Let's Encrypt.
- The origin cert at `/data/nginx-proxy/ssl/occ.io.vn.pem.cloudflare.bak` covers `*.occ.io.vn` and `occ.io.vn`.

### Site still shows old afk-game content
- DNS propagation may take a few minutes (usually < 5 min with Cloudflare).
- Clear your browser cache or test in a private window.
- Check: `curl -sI https://occ.io.vn | grep -i 'server\|cf-ray'` – if you see `CF-Ray:` header, Cloudflare is active.

### afk-game is now unreachable at occ.io.vn
- The afk-game was moved to subdomain `afk-game.occ.io.vn`.  
- Add a DNS A record in Cloudflare: `afk-game` → `51.79.138.246` (Proxied).

---

## 7. Architecture Summary

```
Browser
  │
  │  HTTPS (TLS – Cloudflare edge cert)
  ▼
Cloudflare CDN  (occ.io.vn proxied)
  │
  │  HTTPS (TLS – Cloudflare Origin Cert, Full Strict)
  ▼
VPS 51.79.138.246:443
  │
  │  nginx (travian-proxy Docker container)
  │  /data/nginx-proxy/conf/typing-for-kids.conf
  │
  │  HTTP proxy → 172.17.0.1:11100
  ▼
KidTyping VN app (host process, systemd: kidtyping.service)
  /data/typing-for-kids/kidtyping-vn-linux
```

---

## 8. Re-deployment (future updates)

From your local machine after pushing to `git@github.com:keyhobbit/typing-for-kids.git` branch `master`:

```bash
cd /home/dragonshit/projects/typing-for-kids

# 1. Build static binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o kidtyping-vn-linux .

# 2. Upload to VPS
scp kidtyping-vn-linux vps:/data/typing-for-kids/kidtyping-vn-linux

# 3. Update static assets / templates if changed
rsync -avz --delete static/ vps:/data/typing-for-kids/static/
rsync -avz --delete templates/ vps:/data/typing-for-kids/templates/

# 4. Restart service
ssh vps "systemctl restart kidtyping && systemctl status kidtyping --no-pager"
```
