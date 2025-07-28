# MWP (Mikrotik WireGuard Portal)

A modern web portal for managing Mikrotik WireGuard VPN peers, built with **Go**, **React**, and **TypeScript**. Easily
create, update, share, and monitor WireGuard peers with a user-friendly interface and robust backend.

---

## Table of Contents

- [Features](#features)
- [Screenshots](#screenshots)
- [Getting-Started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Backend Setup](#backend-setup)
    - [Frontend Setup](#frontend-setup)
- [Configuration](#configuration)
- [Usage](#usage)
- [API Reference](#api-reference)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)
- [Contact](#contact)

---

## Features

- 🚀 Create, update, and delete WireGuard peers
- 🔒 Secure key generation and management
- 📊 Real-time peer status and usage statistics
- ⏱️ Automatic peer expiration time (TTL)
- Traffic Limitations and auto-expiration
- Bandwidth Limitation for upload and download
- 📤 Share peer configs via secure links and QR codes
- 🛠️ Mikrotik RouterOS API integration
- Multi-Server support (for future)
- 🔑 API token authentication
- 🖥️ Responsive React + Golang backend

---

## Getting Started

### Prerequisites

- **Go** 1.20 or later
- **Node.js** 18 or later
- **pnpm**
- Mikrotik RouterOS device with API access
- PostgreSQL or SQLite (depending on backend config)

---

### Project Setup

```bash
git clone https://github.com/maahdima/mwp.git
cd mwp

cp .env.example .env
# Edit .env with your Mikrotik credentials and DB settings

go mod tidy
go run main.go
```

---

Visit `http://localhost:3000` in your browser.

---

## Configuration

### Backend `.env`

```env
API_PORT=8080
MIKROTIK_HOST=192.168.88.1
MIKROTIK_PORT=8728
MIKROTIK_USER=admin
MIKROTIK_PASSWORD=yourpassword
DB_URL=postgres://user:pass@localhost:5432/mwp?sslmode=disable
```

---

## Usage

1. **Login** using Mikrotik credentials
2. **Create a new peer** from the dashboard
3. **Share config** via QR code or secure link
4. **Monitor stats** such as last handshake, data usage
5. **Auto-expire** peers after defined TTL
6. **Revoke or edit** existing peers

---

## Configuration

MWP is configured using environment variables. Create a `.env` file in the root project directory (for Docker) or in the
`api` directory (for manual setup).

| Variable                | Description                                          | Default                 | Required |
|-------------------------|------------------------------------------------------|-------------------------|----------|
| `MIKROTIK_HOST`         | The IP address or domain of your Mikrotik router.    | `""`                    | Yes      |
| `MIKROTIK_PORT`         | The API port for your Mikrotik router.               | `8728`                  | Yes      |
| `MIKROTIK_USER`         | The username for the Mikrotik API user.              | `""`                    | Yes      |
| `MIKROTIK_PASSWORD`     | The password for the Mikrotik API user.              | `""`                    | Yes      |
| `MIKROTIK_WG_INTERFACE` | The name of the WireGuard interface on your router.  | `wireguard1`            | Yes      |
| `SERVER_PORT`           | The port for the backend Go API server.              | `8080`                  | No       |
| `JWT_SECRET`            | A secret string for signing authentication tokens.   | `your-secret-key`       | Yes      |
| `WEB_URL`               | The public URL of the frontend for generating links. | `http://localhost:3000` | No       |
| `ADMIN_USER`            | The username for the portal's admin account.         | `admin`                 | No       |
| `ADMIN_PASSWORD`        | The password for the portal's admin account.         | `admin`                 | No       |

---

## Roadmap

- [x] Mikrotik WireGuard peer management
- [x] QR code + link sharing for peer configs
- [x] Automatic TTL + traffic expiration
- [x] Stats dashboard (last handshake, usage)
- [x] Mikrotik backup/sync mechanism
- [x] API token authentication
- [x] Theme customization (light/dark)
- [x] Docker support
- [x] Single Binary Build
- [ ] Telegram Bot support for notifications
- [ ] Multi-server support

---

## Screenshots

### Dashboard-1

![Dashboard](docs/screenshots/dashboard-1.png)

### Dashboard-2

![Dashboard](docs/screenshots/dashboard-2.png)

### Peers Page

![Peers Page](docs/screenshots/peers.png)

### Craete Peer

![Peers Page](docs/screenshots/create-peer.png)

### Share Peer

![Share Peer](docs/screenshots/share.png)
---

## Contributing

Contributions are welcome! 🎉

1. Fork the repo
2. Create a feature branch:
   ```bash
   git checkout -b feat/my-feature
   ```
3. Commit your changes:
   ```bash
   git commit -am "Add my feature"
   ```
4. Push to the branch:
   ```bash
   git push origin feat/my-feature
   ```
5. Open a pull request

> For major changes, please open an issue first to discuss what you’d like to change.

---

## License

This project is licensed under the [MIT License](LICENSE).

---

## Contact

Created by [Maahdima](https://github.com/maahdima) – feel free to reach out via GitHub issues or pull requests.