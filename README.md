# Tesla Fleet API SDK Command Demo (Go)

Send commands to your Tesla using the official Vehicle Command SDK. Works locally with just your credentials - no AWS, no extra services needed.

**Supported commands:**
- 🔒 Lock/Unlock doors
- 🛡️ Sentry Mode on/off
- And more via Tesla Fleet API

> ⚠️ **Security Note:** Never commit `.env`, `tesla-tokens.json`, or files in `config/` to git - they contain secrets and private keys. These are already in `.gitignore`.

---

## ⚡ Quick Start (10 minutes)

### Prerequisites
- **Tesla Fleet API App** - [Set up at developer.tesla.com](https://developer.tesla.com)
  - Get your `CLIENT_ID` and `CLIENT_SECRET`
- **Go 1.23+** - [Download here](https://golang.org/dl)
- **Docker & Docker Compose** - [Download here](https://www.docker.com/products/docker-desktop)
- **Your Vehicle VIN** - Find it in Tesla: Settings > Software

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
TESLA_VEHICLE_VIN=5YJ3E1EA1KF123456
```

Find your VIN in your Tesla vehicle:
- **Settings > Software > VIN**

### Step 2: Set Up Proxy (One Time)

```bash
bash setup-proxy.sh
```

This generates:
- TLS certificates (`config/tls-cert.pem`, `config/tls-key.pem`)
- Fleet key configuration (`config/fleet-key.pem`)

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
1. Checks for saved OAuth tokens
2. If none found, opens browser for Tesla login
3. You approve the app
4. Copy the authorization **code** from redirect URL
5. Paste code into terminal
6. Tokens saved to `tesla-tokens.json` ✅
7. Interactive menu appears

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

That's it! Your vehicle commands work directly from your laptop.

---

## � Alternative: Kubernetes Deployment (Kind + Helm)

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
3. Log in to your Tesla account
4. Approve the application
5. Copy authorization code from redirect
6. Paste into terminal
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
├── tesla-tokens.json        ← OAuth tokens (auto-generated)
├── private-key.pem          ← Your private key for signing
├── public-key.pem           ← Public key (shared with vehicle)
├── config/                  ← Auto-generated by setup-proxy.sh
│   ├── tls-cert.pem         ← TLS certificate
│   ├── tls-key.pem          ← TLS private key
│   └── fleet-key.pem        ← Copy of private key
├── go.mod                   ← Go dependencies
├── README.md                ← This file
└── SETUP.md                 ← Quick reference
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

## 🚀 Usage Examples

**Basic usage (interactive menu):**
```bash
go run tesla-cli.go
```

**To stop the proxy:**
```bash
docker-compose down
```

**View proxy logs:**
```bash
docker-compose logs -f tesla-proxy
```

**Force new authentication:**
```bash
rm tesla-tokens.json
go run tesla-cli.go
```

---

## 💻 Add More Commands

Edit `tesla-cli.go` and add new menu options. Example:

```go
case "6":
    fmt.Println("🚪 Opening trunk...")
    if err := sendCommand(vehicleID, tokens.AccessToken, "open_trunk", proxyURL, map[string]interface{}{}); err != nil {
        fmt.Printf("❌ Error: %v\n", err)
    } else {
        fmt.Println("✅ Trunk opened!")
    }
```

See `sendCommand()` function in `tesla-cli.go` for the pattern.

---

## 🔒 Security

- ✅ **End-to-end encryption** with your vehicle
- ✅ **Cryptographic signing** with your private key
- ✅ **OAuth 2.0** authentication
- ✅ **Self-signed TLS** for local proxy (safe for localhost)
- ✅ **No plaintext credentials** in logs
- ✅ **Token expiration** handled automatically

---

## 📖 References

- [Tesla Vehicle Command SDK](https://github.com/teslamotors/vehicle-command) - Official Go SDK
- [Tesla Fleet API Docs](https://developer.tesla.com/docs/fleet-api) - Complete API reference
- [Vehicle Command Protocol](https://github.com/teslamotors/vehicle-command/blob/main/README.md) - Technical details

---

## ✅ Current Status

**Working:**
- ✅ OAuth 2.0 browser-based authentication
- ✅ Token caching and reuse
- ✅ Lock/Unlock doors
- ✅ Sentry mode control
- ✅ Interactive menu interface
- ✅ Cryptographically signed commands
- ✅ End-to-end vehicle encryption
- ✅ Docker containerized proxy

**Future Ideas:**
- [ ] Add climate, charging, trunk commands
- [ ] Support multiple vehicles
- [ ] REST API wrapper
- [ ] Web dashboard
- [ ] Command scheduling

---

## 🤝 Contributing

Found a bug? Have improvements? Open an issue or PR!

---

## 📄 License

MIT
  - Displays interactive command menu to test lock, unlock, and sentry mode
  - No AWS or DynamoDB required

**Typical output:**
```
🔧 Tesla Command SDK Integration Test (Local Mode)
Vehicle ID: 3744191383554437

Which command would you like to simulate sending?
1. Lock Doors
2. Unlock Doors
3. Set Sentry Mode
4. Quit
Enter choice [1-4]: 
```

## Example Code: Sending Lock/Unlock Commands

Example TypeScript code for SDK usage:
```typescript
import { VehicleCommandSDK } from '@teslamotors/vehicle-command';

const sdk = new VehicleCommandSDK({
  vehicleId: '<your-vehicle-id>',
  accessToken: '<access-token>',
  refreshToken: '<refresh-token>'
});

// Unlock the vehicle
await sdk.unlockDoors();

// Lock the vehicle
await sdk.lockDoors();
```

---

## Build and Run the Go App

To build:
```bash
go build -o tesla-cmd-demo test-tesla-command-sdk-local.go
```
To run:
```bash
VEHICLE_ID=<your-vehicle-id> ./tesla-cmd-demo
```

---

## Before/After Migration Example

**Before (Deprecated REST API):**
```typescript
// backend/src/tesla/tesla-api.ts
async setSentryMode(vehicleId: string, enable: boolean) {
  return this.makeRequest('POST', `/api/1/vehicles/${vehicleId}/command/set_sentry_mode`, {
    on: enable
  })
  // Returns 403: "Tesla Vehicle Command Protocol required"
}
```

**After (SDK):**
```typescript
// backend/src/tesla/tesla-api.ts
async setSentryMode(vehicleId: string, enable: boolean) {
  const sdk = this.initializeSDK(vehicleId)
  return sdk.setSentryMode(enable)
  // Works! Returns success response
}

private initializeSDK(vehicleId: string) {
  return new TeslaVehicleCommandSDK({
    vehicleId,
    accessToken: this.tokens.accessToken,
    refreshToken: this.tokens.refreshToken
  })
}
```

---

## Commands Ready to Implement

All 14 commands are ready immediately after SDK migration:
```json
[
  { "name": "SetSentryMode", "params": "enable: boolean" },
  { "name": "LockDoors", "params": "none" },
  { "name": "UnlockDoors", "params": "none" },
  { "name": "StartCharging", "params": "none" },
  { "name": "StopCharging", "params": "none" },
  { "name": "SetChargeLimit", "params": "percent: number" },
  { "name": "SetClimateTemperature", "params": "temp: number" },
  { "name": "TurnOnClimate", "params": "none" },
  { "name": "TurnOffClimate", "params": "none" },
  { "name": "OpenTrunk", "params": "type: 'front'|'rear'" },
  { "name": "CloseTrunk", "params": "type: 'front'|'rear'" },
  { "name": "OpenWindow", "params": "percent: number" },
  { "name": "CloseWindow", "params": "none" },
  { "name": "UpdateChargeLimit", "params": "percent: number" }
]
```

---

## Troubleshooting

### Vehicle Command Protocol Error (HTTP 403)

**Problem**: When sending commands, you receive:
```
❌ Error: API returned status 403: Tesla Vehicle Command Protocol required
```

**Why**: Tesla deprecated the REST API command endpoint and now requires the **Vehicle Command Protocol**, which uses cryptographic signing to authenticate commands instead of simple Bearer tokens.

**Solution: Use the vehicle-command HTTP Proxy (Recommended) ⭐**

Tesla provides an official HTTP proxy that handles all protocol signing for you automatically:

1. **Install Node.js** (if you don't have it):
   - Download from https://nodejs.org/
   - Or use `brew install node` on macOS, `choco install nodejs` on Windows

2. **Install the vehicle-command proxy**:
   ```bash
   npm install -g @teslamotors/vehicle-command
   ```

3. **Start the proxy in a new terminal**:
   ```bash
   vehicle-command http-server --port 3000
   ```
   Output should show:
   ```
   Server running on http://localhost:3000
   ```

4. **Update the Go code** to use the proxy instead of direct Tesla API:
   - Open `test-tesla-command-sdk-local.go`
   - Find the line: `url := fmt.Sprintf("https://fleet-api.prd.na.vn.cloud.tesla.com/api/1/vehicles/%s/command/%s", vehicleID, command)`
   - Change it to: `url := fmt.Sprintf("http://localhost:8080/api/1/vehicles/%s/command/%s", vehicleID, command)`

5. **Run the program again**:
   ```bash
   go run test-tesla-command-sdk-local.go
   ```

6. **Send commands** - they should now work! The proxy handles all the protocol signing automatically.

**Example flow:**
```
Your Program → HTTP Request → vehicle-command Proxy → Protocol Signing → Tesla API → Vehicle
```

---

### Demo Issues
- **Problem**: Command not found
  ```
bash: go: command not found
  ```
  **Solution**: Install Go from https://golang.org/dl

- **Problem**: OAuth code exchange fails
  ```
Token request failed: ...
  ```
  **Solutions:**
  1. Verify CLIENT_ID and CLIENT_SECRET are correct
  2. Verify REDIRECT_URI matches what's registered in your Tesla app
  3. Check that the authorization code hasn't expired (they expire quickly)

- **Problem**: Browser doesn't open automatically
  ```
  No browser opened, but URL is printed
  ```
  **Solution**: Manually copy and paste the printed URL into your browser

---

## Environment Variables

Optional - if you create a `.env` file, the program will use it:
```bash
TESLA_CLIENT_ID=your_client_id
TESLA_CLIENT_SECRET=your_client_secret
TESLA_REDIRECT_URI=http://localhost:8080/callback
VEHICLE_ID=your_vehicle_id
```

If `.env` is not present, the program will prompt you for these values interactively.

---

## Progress Checklist

- [x] Interactive demo script created
- [x] Token setup integrated (no AWS required)
- [x] Command menu implemented (Lock, Unlock, Set Sentry Mode)
- [x] Documentation complete
- [ ] SDK implementation in backend (your next step)
- [ ] All 14 commands tested
- [ ] Staging deployment
- [ ] Production deployment

---

## Next Steps

### Immediate (Right Now)
```bash
# 1. Run the interactive demo
VEHICLE_ID=<your-vehicle-id> go run test-tesla-command-sdk-local.go

# 2. Follow the prompts to set up OAuth tokens
# 3. Select commands from the menu to test
```

### Soon (This Week)
1. Integrate the SDK into your backend
2. Implement actual lock/unlock/sentry mode commands
3. Connect to real Tesla vehicles for testing
4. Write unit tests

### Later (After Implementation)
1. Deploy to staging
2. Test with real vehicle
3. Deploy to production

---

## Architecture Overview

```
Current (Broken):
  Admin UI → Backend Handler → HTTP REST API → 403 Error ❌

After SDK Migration:
  Admin UI → Backend Handler → Vehicle Command SDK → Vehicle ✅
                              ↓
                       Local Tokens (JSON file)
```

---

## Resources

- **Tesla Vehicle Command SDK**: https://github.com/teslamotors/vehicle-command
- **SDK Documentation**: Included in repository
- **AWS DynamoDB Docs**: https://docs.aws.amazon.com/dynamodb/
- **This Project's Backend**: [backend/README.md](../backend/README.md)
- **API Configuration**: [docs/API_CONFIGURATION.md](../docs/API_CONFIGURATION.md)

---

**Last Updated**: January 31, 2026  
**Status**: Ready for SDK Migration  
**Estimated Implementation Time**: 2-3 hours

**Happy Hacking!**
