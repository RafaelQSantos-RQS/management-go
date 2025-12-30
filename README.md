<p align="center">
  <img src="assets/banner.png" alt="Docker Cleanup Banner" width="100%">
</p>

# 🧹 Docker Cleanup Tool

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![Docker Compliant](https://img.shields.io/badge/Docker-OCI--Compliant-2496ED?style=for-the-badge&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

An automated, lightweight, and professional Go utility designed to keep your Docker environment lean and efficient. It safely removes unused resources without interrupting your active containers.

---

## ✨ Key Features

- **✅ Container Management:** Automatically identifies and removes stopped containers.
- **✅ Volume Pruning:** Safely reclaims disk space by deleting orphan volumes.
- **✅ Network Cleanup:** Removes unused Docker networks.
- **✅ Image Optimization:** Prunes dangling and unused images (while preserving those in use).
- **✅ Daemon Mode:** Can run as a persistent background service with configurable intervals.
- **✅ Cloud Ready:** Fully Dockerized and OCI-compliant.

---

## 🚀 Getting Started

### 🐳 Option 1: Docker (Recommended)
The most portable way to run the tool. Ideal for servers and CI/CD environments.

**One-shot run:**
```bash
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock rafaelqsantos/docker-cleanup
```

**Running as a Daemon (Background Service):**
```bash
docker run -d \
  --name docker-cleanup \
  -e CLEANUP_INTERVAL=24h \
  -v /var/run/docker.sock:/var/run/docker.sock \
  rafaelqsantos/docker-cleanup
```

### 💻 Option 2: Pre-compiled Binaries
Download the latest binaries for your platform from the [Releases](https://github.com/rafael-qsantos/Management-go/releases) page.

```bash
# Example for Linux
./docker-cleanup-linux-amd64
```

### 🛠️ Option 3: Developer Setup
If you prefer running from source:

```bash
git clone https://github.com/rafael-qsantos/Management-go.git
cd Management-go
go run main.go
```

---

## ⚙️ Configuration

Control the behavior of the tool using environment variables:

| Variable | Description | Default | Example |
| :--- | :--- | :--- | :--- |
| `CLEANUP_INTERVAL` | Time to wait between cleanup cycles (Daemon Mode) | Empty (Runs once) | `1h`, `24h`, `30m` |
| `DOCKER_HOST` | Docker socket path (inherited from environment) | `unix:///var/run/docker.sock` | - |

---

## 🤖 CI/CD Automation

This project is fully automated via GitHub Actions:
- **Automated Releases:** Every version tag (`v*`) triggers a multi-platform binary build.
- **Docker Publishing:** Every release is automatically pushed to Docker Hub with `latest` and `version` tags.

---

## 📝 License

Distributed under the MIT License. See `LICENSE` for more information.

---

<p align="center">
  Built with ❤️ by <b>Rafael Queiroz Santos</b>
</p>
