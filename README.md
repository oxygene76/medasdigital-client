# MedasDigital Client - Planet 9 Analysis

A distributed astronomical analysis client for the search of Planet 9 and other trans-Neptunian objects (TNOs) using the MedasDigital blockchain infrastructure.

## 🌌 Overview

The MedasDigital Client is a specialized tool designed to perform distributed astronomical data analysis for the ongoing search for Planet 9 and other distant objects in our solar system. Built on the MedasDigital blockchain, it enables researchers to contribute computational resources while maintaining transparent, verifiable, and immutable records of analysis results.

## 🎉 Current Status - **BLOCKCHAIN INFRASTRUCTURE COMPLETE**

### ✅ **Fully Functional - Blockchain Integration**
- **Real blockchain transactions**: Successfully tested on MedasDigital mainnet
- **Client registration**: Working registration system with permanent client IDs  
- **Key management**: Full key lifecycle management (create, recover, list, delete)
- **Account verification**: Comprehensive blockchain account status checking
- **Local storage**: Automatic backup of registrations in JSON format

**Latest Success**: Client ID `client-071be41f` successfully registered on block 3,565,770
Transaction: [393AAB8CA6273CEC7CCC49A3AB8D3E81E133329267322D0A5799CF7AF5FE55EF](https://explorer.medas-digital.io:3100/medasdigital/tx/393AAB8CA6273CEC7CCC49A3AB8D3E81E133329267322D0A5799CF7AF5FE55EF)

### 🚧 **In Development - Scientific Analysis**
- **Orbital dynamics analysis**: Framework planned, not yet implemented
- **Photometric processing**: Data structures designed, processing pipeline in development  
- **GPU acceleration**: Infrastructure ready, analysis algorithms pending
- **AI/ML integration**: Architecture planned, models not yet trained
- **Survey data processing**: Format support planned, automation pending

**Current Focus**: Building the scientific analysis pipeline on top of the proven blockchain foundation.

## 🔬 Scientific Background

Planet 9 is a hypothetical planet in the outer reaches of our solar system, first proposed based on the unusual clustering of orbits of several trans-Neptunian objects (TNOs). This client provides the foundation infrastructure for distributed analysis of astronomical data.

This client currently provides:
- **✅ Blockchain Foundation**: Complete client registration and identity management
- **✅ Network Integration**: Full MedasDigital blockchain connectivity
- **✅ Capability Declaration**: Register your intended analysis capabilities
- **🚧 Analysis Pipeline**: Scientific algorithms in active development
- **🚧 Data Processing**: Astronomical data formats and processing (planned)
- **🚧 GPU Acceleration**: Hardware integration framework (in progress)

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

This creates `~/.medasdigital-client/config.yaml`:

```yaml
chain:
  chain_id: medasdigital-2
  rpc_endpoint: https://rpc.medas-digital.io:26657
  bech32_prefix: medas
  base_denom: umedas
client:
  keyring_dir: /home/user/.medasdigital-client/keyring
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

### Create Your Identity

```bash
# Create a new key (generates mnemonic)
./bin/medasdigital-client keys add MyAnalysisNode

# Or recover from existing mnemonic
./bin/medasdigital-client keys add MyAnalysisNode --recover

# List all keys
./bin/medasdigital-client keys list

# Show specific key details
./bin/medasdigital-client keys show MyAnalysisNode
```

## 📊 Usage

### 1. Register Your Analysis Node

```bash
# Register on the blockchain with default capabilities
./bin/medasdigital-client register --from MyAnalysisNode

# Register with custom capabilities and metadata
./bin/medasdigital-client register \
  --from MyAnalysisNode \
  --capabilities orbital_dynamics,photometric_analysis,ai_training \
  --metadata "Institution: Your Observatory, Location: Chile"

# Register with manual gas limit
./bin/medasdigital-client register \
  --from MyAnalysisNode \
  --gas 200000
```

**Successful Registration Output:**
```
🎉 CLIENT SUCCESSFULLY REGISTERED ON BLOCKCHAIN!
===================================================
🆔 Client ID: client-071be41f
📍 Address: medas1y0n5v8m0jn0mp37vp74qjq3nfk7zf6ahsgywwr
⛓️  Chain: medasdigital-2
🔧 Capabilities: [orbital_dynamics photometric_analysis]
📊 Transaction Hash: 393AAB8CA6273CEC7CCC49A3AB8D3E81E133329267322D0A5799CF7AF5FE55EF
🏔️  Block Height: 3565770
💾 Registration saved to: ~/.medasdigital-client/registrations/
===================================================
```

### 2. Verify Your Account

```bash
# Check account status on blockchain
./bin/medasdigital-client check-account --from MyAnalysisNode

# Or check specific address
./bin/medasdigital-client check-account medas1y0n5v8m0jn0mp37vp74qjq3nfk7zf6ahsgywwr
```

### 3. Monitor Status

```bash
# Check client status
./bin/medasdigital-client status

# Check blockchain connection
./bin/medasdigital-client check-account --from MyAnalysisNode
```

## 📁 Local Storage

### Registration Files

Your registrations are automatically saved locally:

```bash
# View your registration
cat ~/.medasdigital-client/registrations/registration-client-071be41f.json

# View all registrations index
cat ~/.medasdigital-client/registrations/index.json

# List all registration files
ls -la ~/.medasdigital-client/registrations/
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

## 🔧 Advanced Usage

### Custom Configuration

```bash
# Use custom config file
./bin/medasdigital-client --config /path/to/config.yaml register --from MyKey

# Use custom home directory
./bin/medasdigital-client --home /path/to/data register --from MyKey

# Use different keyring backend
./bin/medasdigital-client register --from MyKey --keyring-backend file
```

### Development Mode

```bash
# Build with debug info
make build-debug

# Run with verbose logging
./bin/medasdigital-client register --from MyKey --verbose

# Test configuration
./bin/medasdigital-client config validate
```

## 🌐 Network Configuration

### Mainnet (Default)
- **Chain ID**: `medasdigital-2`
- **RPC Endpoint**: `https://rpc.medas-digital.io:26657`
- **Explorer**: `https://explorer.medas-digital.io`
- **Denomination**: `umedas` (micro-medas)

### Gas and Fees
- **Recommended Gas**: 200,000 for registration
- **Fee Calculation**: Automatic (minimum 5,000 umedas)
- **Typical Cost**: ~0.005 MEDAS per registration

## 🚀 Future Features (Roadmap)

### Phase 1: Foundation ✅ **COMPLETED**
- [x] Blockchain integration with Cosmos SDK v0.50.10
- [x] Client registration system with permanent on-chain storage
- [x] Comprehensive key management and security
- [x] Local storage and backup systems
- [x] Account verification and status checking
- [x] Gas optimization and fee calculation

### Phase 2: Analysis Framework 🚧 **IN DEVELOPMENT**
- [ ] Orbital dynamics analysis pipeline
- [ ] Photometric data processing modules
- [ ] GPU acceleration integration (CUDA/OpenCL)
- [ ] AI/ML model framework integration
- [ ] Distributed computation coordination
- [ ] Data format standardization (FITS, CSV, JSON)

### Phase 3: Advanced Analytics 📋 **PLANNED**
- [ ] Real-time survey data processing
- [ ] Collaborative analysis verification system
- [ ] Multi-institutional data sharing protocols
- [ ] Publication-ready result generation
- [ ] Integration with major observatories and surveys

## 🛠️ Development

### Build from Source

```bash
# Clone repository
git clone https://github.com/oxygene76/medasdigital-client.git
cd medasdigital-client

# Install dependencies
go mod tidy

# Run tests
go test -v ./...

# Build
make build

# Install locally
make install
```

### Project Structure

```
medasdigital-client/
├── cmd/medasdigital-client/    # Main application
├── pkg/
│   ├── client/                 # Client logic
│   ├── blockchain/             # Blockchain integration
│   ├── analysis/               # Analysis algorithms
│   └── gpu/                    # GPU acceleration
├── internal/types/             # Internal types
├── Makefile                    # Build configuration
├── go.mod                      # Go dependencies
└── README.md                   # This file
```

## 🔐 Security

- **Keyring Security**: Uses Cosmos SDK keyring with multiple backend options
- **Transaction Signing**: All transactions are cryptographically signed
- **Local Storage**: Registration data backed up locally in JSON format
- **Blockchain Verification**: All registrations are permanently stored on-chain

## 🐛 Troubleshooting

### Common Issues

**Key not found:**
```bash
# Create a key first
./bin/medasdigital-client keys add MyAnalysisNode
```

**Insufficient funds:**
```bash
# Check account status
./bin/medasdigital-client check-account --from MyAnalysisNode
# Ensure your account has umedas tokens for transaction fees
```

**Connection issues:**
```bash
# Test blockchain connection
curl -s https://rpc.medas-digital.io:26657/status
```

**Build issues:**
```bash
# Clean and rebuild
make clean
go mod tidy
make build
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🌟 Acknowledgments

- **Mike Brown** and **Konstantin Batygin** for the original Planet 9 hypothesis
- **Minor Planet Center** for maintaining the TNO database
- **Cosmos SDK** for blockchain infrastructure
- **MedasDigital** for the blockchain network

## 📞 Contact

- **Project Repository**: https://github.com/oxygene76/medasdigital-client
- **Technical Support**: support@medas-digital.io
- **Scientific Inquiries**: science@medas-digital.io
- **Blockchain Explorer**: https://explorer.medas-digital.io:3100

---

**🚀 Ready to join the network infrastructure? Get started in 3 commands:**

```bash
make build
./bin/medasdigital-client init
./bin/medasdigital-client keys add MyAnalysisNode
./bin/medasdigital-client register --from MyAnalysisNode
```

*Join the distributed blockchain network for Planet 9 research - register your node and prepare for upcoming scientific analysis capabilities!*

**📊 Current Capabilities**: Blockchain client registration and network participation  
**🔬 Coming Soon**: Full astronomical analysis pipeline and GPU-accelerated computations
