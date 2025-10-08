# MedasDigital Client v2.0

A blockchain-based distributed computing client for the Medas Digital network, featuring smart contract v2.0 with enhanced security and reliability features for computation providers.

## 🌌 Overview

The MedasDigital Client is a comprehensive tool for participating in the MedasDigital distributed computing network. It enables both computation providers to offer services and clients to submit jobs, all managed through blockchain smart contracts with transparent, verifiable, and immutable records.

## 🎉 Current Status - **v2.0 FEATURES COMPLETE**

### ✅ **Smart Contract v2.0 - Production Ready**
- **Heartbeat Monitoring**: Providers stay active with 24-hour timeout protection
- **Job Management**: Submit, cancel, and fail jobs with automatic refunds
- **WebSocket Stability**: Auto-reconnection prevents provider crashes
- **Timeout Processing**: Jobs auto-expire after 1 hour with full refund
- **Balance Harvesting**: Automatic fund management for providers

### ✅ **Blockchain Infrastructure - Proven**
- **Real blockchain transactions**: Successfully tested on MedasDigital mainnet
- **Client registration**: Working registration system with permanent client IDs
- **Key management**: Full key lifecycle management (create, recover, list, delete)
- **Account verification**: Comprehensive blockchain account status checking
- **Smart Contract**: v2.0 deployed at `medas1xr3rq8yvd7qplsw5yx90ftsr2zdhg4e9z60h5duusgxpv72hud3s3cca97`

## 🚀 Features

### Provider Features (v2.0)
- **Automatic Heartbeat**: Keep provider active without manual intervention
- **WebSocket Reconnection**: No more crashes after network issues
- **Job Processing**: Automatic job assignment via WebSocket
- **Failed Job Handling**: Mark jobs as failed with automatic refunds
- **Balance Management**: Auto-harvest excess funds to funding address
- **HTTP Health API**: Monitor provider status via REST endpoints

### Client Features
- **Job Submission**: Submit computing jobs with automatic provider matching
- **Job Cancellation**: Cancel within 5 minutes for full refund
- **Status Tracking**: Monitor job progress and results
- **Multiple Job Types**: PI calculation, orbital dynamics, photometric analysis

### Security Features
- **Provider Inactivity Detection**: 24-hour heartbeat timeout
- **Job Timeouts**: 1-hour automatic expiry with refund
- **Emergency Pause**: Contract-wide pause capability
- **Rate Limiting**: Protection for free test services
- **Secure WebSocket**: Ping/pong keep-alive mechanism

## 🛠️ Installation & Build

### Prerequisites
- **Go 1.21** or later
- **Git** for cloning the repository
- **Make** for building
- **Access** to MedasDigital blockchain RPC endpoint

### Quick Start

```bash
# Clone the repository
git clone https://github.com/oxygene76/medasdigital-client.git
cd medasdigital-client

# Install dependencies
go mod download

# Build the client
make build

# Verify build
./bin/medasdigital-client --version
```

### Build Commands

```bash
# Build optimized binary
make build

# Build with debug information
make build-debug

# Run tests
make test

# Clean build artifacts
make clean

# Install to system PATH
make install
```

## ⚙️ Configuration

### Initialize Client

```bash
# Initialize configuration
./bin/medasdigital-client init
```

### Provider Configuration

Edit `~/.medasdigital-client/config.yaml`:

```yaml
chain:
    chain_id: medasdigital-2
    rpc_endpoint: https://rpc.medas-digital.io:26657
    bech32_prefix: medas
    base_denom: umedas

provider:
    enabled: true
    key_name: "test-provider"
    keyring_backend: "test"
    funding_address: "medas1kc7lctfazdpd8y6ecapdfv3d6ch97prc58qaem"
    min_balance: 50000000
    max_balance: 100000000
    endpoint: "https://provider.medas-digital.io:8080"
    port: 8080
    workers: 4
    harvest_interval_hours: 1
    heartbeat_interval_minutes: 360  # 6 hours recommended

contract:
    address: medas1xr3rq8yvd7qplsw5yx90ftsr2zdhg4e9z60h5duusgxpv72hud3s3cca97

client:
    keyring_dir: /root/.medasdigital/
    keyring_backend: test
    capabilities:
        - orbital_dynamics
        - photometric_analysis

gpu:
    enabled: false
    device_id: 0
    memory_limit: 8192
```

## 🔑 Key Management

```bash
# Create a new key (generates mnemonic)
./bin/medasdigital-client keys add provider-key

# Recover from existing mnemonic
./bin/medasdigital-client keys add provider-key --recover

# List all keys
./bin/medasdigital-client keys list

# Show specific key details
./bin/medasdigital-client keys show provider-key

# Delete a key
./bin/medasdigital-client keys delete provider-key
```

## 📊 Provider Operations

### 1. Register and Start Provider

```bash
# Register provider with contract v2.0
./bin/medasdigital-client contract provider-node --register

# Start provider node (includes all v2.0 features)
./bin/medasdigital-client contract provider-node
```

**Provider v2.0 Output:**
```
=== Provider Node v2.0 ===
Provider Address: medas1f5zg2ju7r4ls988vhlx6sr4nj0l9lpvwgeu0gt
Contract (v2.0): medas1xr3rq8yvd7qplsw5yx90ftsr2zdhg4e9z60h5duusgxpv72hud3s3cca97
Heartbeat: every 360 minutes
✅ Provider registered
🚀 Starting with v2.0 features:
  ✅ Automatic heartbeat every 360 minutes
  ✅ WebSocket auto-reconnection
  ✅ Job failure handling with refunds
  ✅ Balance auto-harvesting
2025/10/08 16:23:43 Provider Node Started (v2.0)
2025/10/08 16:23:43   Name: MEDAS Provider Node v2.0
2025/10/08 16:23:43   Address: medas1f5zg2ju7r4ls988vhlx6sr4nj0l9lpvwgeu0gt
2025/10/08 16:23:43   Endpoint: https://provider.medas-digital.io:8080
2025/10/08 16:23:43   Listening for jobs...
2025/10/08 16:23:43 ✅ WebSocket connected and subscribed
```

### 2. Monitor Provider Health

```bash
# Check on-chain status
./bin/medasdigital-client contract provider-status

# HTTP health check
curl http://localhost:8080/health

# Response includes heartbeat info:
{
  "status": "healthy",
  "provider": "medas1f5zg2ju7r4ls988vhlx6sr4nj0l9lpvwgeu0gt",
  "heartbeat": {
    "last_sent": "2025-10-08T18:04:28Z",
    "seconds_ago": 30,
    "minutes_ago": 0,
    "active": true,
    "next_in": "5h59m30s"
  },
  "websocket_connected": true,
  "reconnect_attempts": 0
}
```

### 3. View Job Results

```bash
# Access completed job results
curl http://localhost:8080/results/pi_calculation-1.json
```

## 💼 Client Operations

### Submit Computing Job

```bash
# Submit PI calculation job
./bin/medasdigital-client contract submit-job \
  --type pi_calculation \
  --digits 1000 \
  --from client-key \
  --payment 1000000umedas
```

### Cancel Job (within 5 minutes)

```bash
./bin/medasdigital-client contract cancel-job \
  --job-id 1 \
  --from client-key
```

### Check Job Status

```bash
./bin/medasdigital-client contract get-job --job-id 1
```

## 🔧 Contract Management

### View Configuration

```bash
./bin/medasdigital-client contract get-config

# Output:
=== Contract Configuration v2.0 ===
Community Pool: medas1kc7lctfazdpd8y6ecapdfv3d6ch97prc58qaem
Community Fee: 15%
Job Timeout: 3600 seconds (60 minutes)
Heartbeat Timeout: 86400 seconds (24 hours)
Contract Paused: false
```

### Manual Operations

```bash
# Send manual heartbeat
./bin/medasdigital-client contract heartbeat --from provider-key

# List all providers
./bin/medasdigital-client contract list-providers
```

## 🌐 Network Information

### Smart Contract v2.0
- **Address**: `medas1xr3rq8yvd7qplsw5yx90ftsr2zdhg4e9z60h5duusgxpv72hud3s3cca97`
- **Community Pool**: `medas1kc7lctfazdpd8y6ecapdfv3d6ch97prc58qaem`
- **Fee Structure**: 85% provider, 15% community

### Timeouts and Limits
- **Job Timeout**: 1 hour (automatic refund)
- **Heartbeat Timeout**: 24 hours (provider marked inactive)
- **Cancel Window**: 5 minutes after submission
- **WebSocket Reconnect**: Max 10 attempts with exponential backoff

### Mainnet Configuration
- **Chain ID**: `medasdigital-2`
- **RPC Endpoint**: `https://rpc.medas-digital.io:26657`
- **Block Explorer**: `https://explorer.medas-digital.io:3100`
- **Denomination**: `umedas` (1 MEDAS = 1,000,000 umedas)

## 📈 Job Lifecycle

1. **Submit**: Client submits job with payment
2. **Assign**: Contract assigns to active provider
3. **Process**: Provider executes computation
4. **Complete/Fail**: Provider submits result or marks failed
5. **Payment**: Automatic distribution (85% provider, 15% community)

### Automatic Safety Features
- ⏱️ Jobs timeout after 1 hour → automatic refund
- 💔 Provider disconnects → job reassigned
- ❌ Job fails → automatic refund minus gas
- 🚫 Provider inactive 24h → removed from active pool

## 🛠️ Free Test Service

Run limited free computation service:

```bash
# Start free service (max 100 digits, rate limited)
./bin/medasdigital-client serve --port 8080

# Direct PI calculation (max 1000 digits for CLI)
./bin/medasdigital-client pi calculate 100
```

## 🐛 Troubleshooting

### Provider Issues

**Not receiving jobs:**
```bash
# Check registration
./bin/medasdigital-client contract list-providers

# Verify heartbeat
curl http://localhost:8080/health | jq .heartbeat

# Check WebSocket in logs
grep "WebSocket connected" provider.log
```

**WebSocket crashes (fixed in v2.0):**
- Auto-reconnection with exponential backoff
- Max 10 attempts before giving up
- Ping/pong keep-alive every 30 seconds

### Balance Management

Enable auto-harvesting:
```yaml
provider:
    funding_address: "medas1..."
    min_balance: 50000000    # Keep minimum
    max_balance: 100000000    # Harvest excess
    harvest_interval_hours: 1
```

### Client Issues

**Insufficient funds:**
```bash
# Check account balance
./bin/medasdigital-client check-account --from client-key

# Fund account
medasdigitald tx bank send <from-wallet> <client-address> 10000000umedas \
  --keyring-backend test \
  --chain-id medasdigital-2 \
  --node https://rpc.medas-digital.io:26657 \
  -y
```

**Key not found:**
```bash
# Create key first
./bin/medasdigital-client keys add client-key

# Or recover from mnemonic
./bin/medasdigital-client keys add client-key --recover
```

## 📁 Local Storage

### Registration Data
```bash
# View registrations
ls ~/.medasdigital-client/registrations/

# View registration details
cat ~/.medasdigital-client/registrations/registration-*.json
```

### Logs and State
```bash
# Provider logs
tail -f ~/.medasdigital-client/provider.log

# View job results
ls ~/.medasdigital-client/results/
```

### Registration Data Format

```json
{
  "transaction_hash": "393AAB8CA6273CEC7CCC49A3AB8D3E81E133329267322D0A5799CF7AF5FE55EF",
  "client_id": "client-071be41f",
  "registration_data": {
    "client_address": "medas1y0n5v8m0jn0mp37vp74qjq3nfk7zf6ahsgywwr",
    "capabilities": ["orbital_dynamics", "photometric_analysis"],
    "metadata": "Institution: Your Observatory",
    "timestamp": "2025-07-12T17:44:20.338470471Z",
    "version": "1.0.0"
  },
  "block_height": 3565770,
  "registered_at": "2025-07-12T17:44:20Z"
}
```

## 🔐 Security

- **Keyring Security**: Cosmos SDK keyring with configurable backend
- **Transaction Signing**: All transactions cryptographically signed
- **WebSocket Security**: TLS encryption with auto-reconnection
- **Rate Limiting**: Built-in protection for public endpoints
- **Job Verification**: On-chain result hashes for verification

## 📄 Version History

### v2.0.0 (Current)
- Smart contract v2.0 with heartbeat system
- WebSocket auto-reconnection (no more crashes)
- Job cancellation and failure handling
- Automatic refunds for timeouts/failures
- Provider health monitoring API
- Balance auto-harvesting
- Emergency pause capability

### v1.0.0
- Initial release
- Basic job submission and processing
- Blockchain registration system
- Key management
- Manual provider management

## 📄 Project Structure

```
medasdigital-client/
├── cmd/medasdigital-client/    # Main application
│   ├── main.go                 # Entry point
│   ├── config.go               # Configuration
│   ├── contract_commands.go    # Contract interactions
│   └── payment_service.go      # Payment processing
├── pkg/
│   ├── client/                 # Client logic
│   ├── blockchain/             # Blockchain integration
│   ├── contract/               # Smart contract interface
│   │   ├── provider.go         # Provider node implementation
│   │   └── types.go            # Contract types
│   ├── compute/                # Computation engines
│   └── analysis/               # Analysis algorithms
├── Makefile                    # Build configuration
├── go.mod                      # Go dependencies
└── README.md                   # This file
```

## 🌟 Acknowledgments

- **Cosmos SDK** team for blockchain infrastructure
- **MedasDigital** community for network support
- **CosmWasm** for smart contract capabilities
- **Mike Brown** and **Konstantin Batygin** for the original Planet 9 hypothesis

## 📞 Support

- **GitHub**: https://github.com/oxygene76/medasdigital-client
- **Chain RPC**: https://rpc.medas-digital.io:26657
- **Documentation**: https://docs.medas-digital.io
- **Support Email**: support@medas-digital.io
- **Block Explorer**: https://explorer.medas-digital.io:3100

---

## 🚀 Quick Start Guide

### For Providers

```bash
# 1. Build and initialize
make build
./bin/medasdigital-client init

# 2. Create provider key and fund it
./bin/medasdigital-client keys add my-provider
# Get address and send MEDAS tokens to it

# 3. Configure provider in config.yaml
# Edit: ~/.medasdigital-client/config.yaml

# 4. Register and start provider
./bin/medasdigital-client contract provider-node --register
```

### For Clients

```bash
# 1. Create client key and fund it
./bin/medasdigital-client keys add my-client
# Get address and send MEDAS tokens to it

# 2. Submit a job
./bin/medasdigital-client contract submit-job \
  --type pi_calculation \
  --digits 1000 \
  --from my-client \
  --payment 1000000umedas

# 3. Check job status
./bin/medasdigital-client contract get-job --job-id 1
```

### Monitor Operations

```bash
# Provider health check
curl http://localhost:8080/health | jq

# Contract status
./bin/medasdigital-client contract get-config

# List all providers
./bin/medasdigital-client contract list-providers

# Check your balance
./bin/medasdigital-client check-account --from my-key
```

---

**Version 2.0** - Production Ready with Enhanced Security and Reliability

*Join the MedasDigital distributed computing network today!*
