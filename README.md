# Tesla Fleet API SDK Command Demo (Go)

Send commands to your Tesla using the official Vehicle Command SDK. Works locally with just your credentials.

**Supported commands:**
- 🔒 Lock/Unlock doors
- 🛡️ Sentry Mode on/off
- Other commands that can be implemented found [here](https://developer.tesla.com/docs/fleet-api/endpoints/vehicle-commands)

> ⚠️ **Security Note:** Never commit `.env`, `tesla-tokens.json`, `.pem`, or files in `config/` to git - they contain secrets and private keys. These are already in `.gitignore`.

---

## ⚡ Quick Start (10 minutes)

### Prerequisites
- **See My Previous Video** [YouTube](https://youtube.com/@quadcitiesteslaguy/playlist/Tesla Development) and follow the instructions in [GitHub](https://github.com/qcteslaguy/tesla-fleet-api-demo/README.md)
- **Tesla Fleet API App** - [Set up at developer.tesla.com](https://developer.tesla.com)
  - Get your `CLIENT_ID` and `CLIENT_SECRET`
- **Your Private and Public Key** - Actually only your private key from when you did the first tutorial. The public key is what you added to your vehicle via the https://tesla.com/_ak/your-url-from-first-tutorial. So as long as the public key still shows in locks you are good, so you only need the private key named as `private-key.pem` which will be copied to `config/fleet-key.pem` to sign the commands via the proxy.
- **Go 1.23+** - [Download here](https://golang.org/dl)
- **Docker & Docker Compose** - [Download here](https://www.docker.com/products/docker-desktop)
- **Your Vehicle VIN** - Find it in the Tesla App at the bottom of the screen.
- **Git Bash or WSL** - A git bash terminal is best for this project but you could also use WSL or convert the .sh script to .bat if you know how.

### Step 1: Configure Credentials

```bash
git clone https://github.com/qcteslaguy/tesla-fleet-api-sdk-commands-demo
cd tesla-fleet-api-sdk-commands-demo
cp .env.example .env
```

**Edit `.env` and add your Tesla credentials:**
```ini
TESLA_CLIENT_ID=your_client_id_from_developer.tesla.com
TESLA_CLIENT_SECRET=your_client_secret
TESLA_REDIRECT_URI=http://localhost:8080/callback
TESLA_VEHICLE_VIN=your_vehicle_vin
```

You can also find your VIN in your Tesla vehicle:
- **Settings > Software > VIN**

Copy your private key from the previous tutorial to the root folder as `private-key.pem`.

### Step 2: Set Up Proxy (One Time)

```bash
bash setup-proxy.sh
```

This generates:
- TLS certificates (`config/tls-cert.pem`, `config/tls-key.pem`)
- Copies the `private-key.pem` to Fleet key configuration (`config/fleet-key.pem`)

### Step 3: Start the Docker Proxy

```bash
docker-compose up -d
```

Verify it's running:
```bash
docker-compose ps
docker-compose logs tesla-proxy
```

### Step 4: Run the Interactive CLI

```bash
go run tesla-cli.go
```

**What happens on first run:**
1. Checks for saved OAuth tokens in `tesla-tokens.json`
2. If none found, uses the browser to get a code
3. Once it gets the cod **should happen automatically**
4. Tokens saved to `tesla-tokens.json` ✅
5. Interactive menu appears

**NOTE:** If you have any errors from above it is likely you didn't add http://localhost:8080/callback as a valid uri on developer.tesla.com for your app.

**Menu:**
```
What would you like to do?
1. Lock Doors
2. Unlock Doors
3. Sentry Mode ON
4. Sentry Mode OFF
5. Quit

Enter choice [1-5]: 
```

That's it! Your vehicle commands work directly from your computer.


**Cleanup:**

Run the followig when done to remove the proxy

```bash
docker-compose down
```
---

## � Alternative: Kubernetes Deployment (Kind + Helm) of a proxy on Windows

For advanced users who want to run the proxy in Kubernetes instead of Docker Compose.

### Prerequisites
```bash
# Install Kind (Kubernetes in Docker)
choco install kind          # Windows with Chocolatey
# or brew install kind      # macOS with Homebrew
# or download from https://kind.sigs.k8s.io/docs/user/quick-start/#installation

# Install kubectl
choco install kubernetes-cli  # Windows with Chocolatey
# or brew install kubectl     # macOS
# or see https://kubernetes.io/docs/tasks/tools/

# Install Helm
choco install kubernetes-helm  # Windows with Chocolatey
# or brew install helm         # macOS
# or see https://helm.sh/docs/intro/install/

# Verify installation
kind version
kubectl version --client
helm version
```

### Deploy to Kind

Run the setup script:
```bash
bash setup-kind-cluster.sh
```

This will:
1. Create a local Kubernetes cluster (Kind)
2. Generate TLS certificates
3. Create a Helm chart
4. Deploy Tesla proxy to the cluster
5. Expose the proxy via kubectl port-forward

### Access the Kubernetes Proxy

```bash
# Forward port from cluster to localhost
kubectl port-forward -n tesla-proxy service/tesla-proxy 4443:4443

# In another terminal, run the CLI
go run tesla-cli.go
```

### Manage the Kubernetes Deployment

```bash
# View cluster status
kubectl get all -n tesla-proxy

# View logs
kubectl logs -n tesla-proxy -l app.kubernetes.io/instance=tesla-proxy

# Delete the cluster
kind delete cluster --name tesla-proxy
```

---

## �🔐 Authentication Flow

**First Time:**
1. Run `go run tesla-cli.go`
2. Browser opens (or URL shown in terminal)
3. Log in to your Tesla account if not logged in first
4. Approve the application if necessary
5. Copy authorization code from redirect or get automatically
6. Paste into terminal or retrieve automatically
7. Done! Tokens saved ✅

**Future Runs:**
- Tokens cached in `tesla-tokens.json`
- Just run `go run tesla-cli.go` and pick a command
- Re-authenticate only if tokens expire (delete `tesla-tokens.json` to force)

---

## 🏗️ Architecture

```
┌─────────────────────┐
│   tesla-cli.go      │  ← You run this
│  (Interactive menu) │
└──────────┬──────────┘
           │
           ↓
┌──────────────────────────┐
│ Docker Proxy             │  ← Runs on localhost:4443
│ (Tesla HTTP Server)      │
│ - TLS encryption         │
│ - Command signing        │
│ - Protocol handling      │
└──────────┬───────────────┘
           │ (HTTPS)
           ↓
┌─────────────────────────┐
│ Tesla Fleet API         │  ← Internet
│ (Tesla Servers)         │
└──────────┬──────────────┘
           │
           ↓
┌─────────────────────────┐
│ Your Tesla Vehicle      │  ← Receives command
└─────────────────────────┘
```

**What Each Component Does:**
- **tesla-cli.go** - Interactive menu, OAuth handling, command execution
- **Docker Proxy** - End-to-end encryption, cryptographic signing, protocol conversion
- **OAuth Tokens** (`tesla-tokens.json`) - Authenticates with Tesla servers
- **Private Key** (`config/fleet-key.pem`) - Signs commands for vehicle authentication
- **TLS Certificates** (`config/tls-*.pem`) - Secure local proxy communication

---

## 📂 Project Structure

```
.
├── tesla-cli.go             ← Main program (run this!)
├── docker-compose.yml       ← Proxy configuration
├── setup-proxy.sh           ← Run once to generate certs
├── .env                     ← Your credentials (create from .env.example)
├── .env.example             ← Template for .env
├── .gitignore               ← What to ignore when pushing to github
├── .gitattributes           ← Git setting to prevent different line returns
├── tesla-tokens.json        ← OAuth tokens (auto-generated by go code)
├── private-key.pem          ← Your private key for signing
├── public-key.pem           ← Public key (shared with vehicle)
├── helm/                    ← Files to deploy to kind with helm
├── config/                  ← Auto-generated by setup-proxy.sh
│   ├── tls-cert.pem         ← TLS certificate
│   ├── tls-key.pem          ← TLS private key
│   └── fleet-key.pem        ← Copy of private key
├── go.mod                   ← Go dependencies
└──README.md                ← This file
```

---

## 📦 Available Commands

Currently in menu:
- **Lock Doors** - `door_lock`
- **Unlock Doors** - `door_unlock`
- **Sentry Mode ON** - `set_sentry_mode` (on: true)
- **Sentry Mode OFF** - `set_sentry_mode` (on: false)

Extend the menu in `tesla-cli.go` to add more commands from the Tesla Fleet API.

---

## 🐛 Troubleshooting

### "TESLA_VEHICLE_VIN not found in .env"
```bash
# Edit .env and add your VIN from Settings > Software
echo "TESLA_VEHICLE_VIN=5YJ3E1EA1KF123456" >> .env
```

### "config/fleet-key.pem not found"
```bash
# Run setup script to generate certificates
bash setup-proxy.sh
```

### "Proxy is unhealthy" or "Connection refused"
```bash
# Check proxy status
docker-compose ps
docker-compose logs -f tesla-proxy

# Restart proxy if needed
docker-compose restart tesla-proxy
```

### "401 Unauthorized"
Token expired. Delete and re-authenticate:
```bash
rm tesla-tokens.json
go run tesla-cli.go
```

### "404 Not Found"
Vehicle not found or unreachable:
- Verify VIN is correct (Settings > Software)
- Ensure vehicle is online in your Tesla app
- Try again in a few seconds

---

## 📖 References

- [Tesla Vehicle Command SDK](https://github.com/teslamotors/vehicle-command) - Official Go SDK
- [Tesla Fleet API Docs](https://developer.tesla.com/docs/fleet-api) - Complete API reference
- [Vehicle Command Protocol](https://github.com/teslamotors/vehicle-command/blob/main/README.md) - Technical details
- [First Tutorial to Get Vehicle Data](https://github.com/qcteslaguy/tesla-fleet-api-demo)

---

## 🤝 Contributing

Found a bug? Have improvements? Open an issue or PR!