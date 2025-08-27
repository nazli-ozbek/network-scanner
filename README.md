# Network Scanner

A container based network scanner that discovers devices in a given IP range using ICMP ping, ARP requests, and hostname resolution. Built with Go, React, SQLite, and Containerlab, the project includes a web based UI and JWT based user authentication system.

## Features

- Scan devices in specified CIDR IP ranges
- Detect online/offline status via ICMP ping
- Resolve MAC addresses (via ARP) and hostnames
- Filter/sort devices by status, hostname, tags, etc.
- Save named IP ranges and scan history
- Secure user login and registration (JWT)
- Containerlab-powered virtual lab simulation

## Tech Stack

- Backend: Go, net/http, SQLite, JWT, bcrypt, swaggo
- Frontend: Next.js (React), Tailwind CSS
- Virtual Lab: Containerlab, Alpine-based simulated devices
- Database: SQLite (with device/user/iprange tables)
- Swagger: for API documentation (/docs endpoint)

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go (if building manually)
- Containerlab:
  ```bash
  curl -sL https://get.containerlab.dev | bash
  ```

### Build the Go Backend (Optional)

```bash
go build -o scanner
./scanner
```

### Running the Containerlab Topology

Run the lab:

```bash
./cmd/run_lab.sh
```

To destroy the lab:

```bash
./cmd/destroy_lab.sh
```

### Accessing the Frontend

Once the lab is up:

- Open browser at: http://localhost:3000
- Register a new account
- Login and begin scanning IP ranges

### Docker Compose (Optional)

If not using Containerlab:

```bash
docker-compose up --build
```

### Default Credentials

This system does not include a default user. You must register via the UI page.

## API Documentation

Once running, open:

```
http://localhost:8080/docs/index.html
```

Swagger is auto-generated using swaggo.
