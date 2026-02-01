# Quick Reference

Complete guide: See [README.md](README.md)

## Option 1: Docker Compose (Recommended for Local Development)

```bash
# 1. Configure credentials
cp .env.example .env
# Edit .env with your CLIENT_ID, CLIENT_SECRET, VEHICLE_VIN

# 2. Generate TLS certificates (one time)
bash setup-proxy.sh

# 3. Start the Docker proxy
docker-compose up -d

# 4. Run the CLI
go run tesla-cli.go
```

## Option 2: Kubernetes (Kind + Helm)

```bash
# 1. Install tools (one time)
choco install kind kubernetes-cli kubernetes-helm

# 2. Configure credentials
cp .env.example .env
# Edit .env with your CLIENT_ID, CLIENT_SECRET, VEHICLE_VIN

# 3. Deploy to Kind cluster
bash setup-kind-cluster.sh

# 4. Port-forward to access proxy
kubectl port-forward -n tesla-proxy service/tesla-proxy 4443:4443

# 5. In another terminal, run the CLI
go run tesla-cli.go
```

## Commands

The interactive menu now has 6 options:

- **Option 1: Wake Vehicle** - Explicitly wake a sleeping vehicle
- **Option 2: Lock Doors** - Lock all doors
- **Option 3: Unlock Doors** - Unlock all doors
- **Option 4: Sentry Mode ON** - Enable Sentry Mode
- **Option 5: Sentry Mode OFF** - Disable Sentry Mode
- **Option 6: Quit** - Exit the CLI

**Note:** Commands automatically wake the vehicle if it's asleep, then retry. No manual intervention needed!

## Troubleshooting

| Issue | Solution |
|-------|----------|
| `config/fleet-key.pem not found` | Run `bash setup-proxy.sh` |
| `No valid tokens` | Delete `tesla-tokens.json`, run `go run tesla-cli.go` |
| Proxy unhealthy | `docker-compose restart tesla-proxy` (Docker) or check `kubectl logs -n tesla-proxy` (Kubernetes) |
| 404 Not Found | Verify VIN in `.env` is correct (must be 17 characters) |
| 401 Unauthorized | Token expired, delete `tesla-tokens.json` |
| `vehicle unavailable: offline or asleep` | Use menu option 1 to wake vehicle, or wake from Tesla app first for faster response |
| Kubernetes extension not seeing cluster | Set kubeconfig in VS Code: `"vs-kubernetes": {"kubeconfig": "\\\\wsl$\\Ubuntu\\home\\plang\\.kube\\config"}` |

## Stop Proxy

```bash
docker-compose down
```

---

👉 Full guide and all commands: [README.md](README.md)

