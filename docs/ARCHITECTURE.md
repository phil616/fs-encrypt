# fs-encrypt Architecture & Specification

## 1. Security Model

### 1.1 Key Derivation
- **Algorithm**: Argon2id
- **Parameters**:
  - Time: 1
  - Memory: 64 MB (64 * 1024 KB)
  - Threads: 4
  - Salt: 16 bytes (Randomly generated per file)
  - Key Length: 32 bytes (256 bits)

### 1.2 Encryption
- **Algorithm**: AES-256-GCM (Galois/Counter Mode)
- **Authentication**: GCM Tag (16 bytes) appended to each chunk.
- **Nonce**: 12 bytes (Randomly generated per chunk).
- **Chunk Size**: 64 KB plaintext blocks.

### 1.3 Threat Model
- **Confidentiality**: Protected by AES-256. Attackers cannot read data without the key.
- **Integrity**: Protected by GCM. Any modification to ciphertext (including reordering chunks) will be detected upon decryption.
- **Side Channels**: Go's `crypto/aes` implementation is generally resistant to timing attacks on modern CPUs (AES-NI). Argon2 is memory-hard to resist GPU cracking.

## 2. File Format Specification

The output file consists of a header followed by a sequence of encrypted chunks.

### 2.1 Header
| Offset | Size | Description |
|--------|------|-------------|
| 0      | 16   | Salt (Used for Key Derivation) |

### 2.2 Chunk Structure
Each chunk corresponds to up to 64KB of compressed data.

| Offset | Size | Description |
|--------|------|-------------|
| 0      | 4    | **Length** (uint32, Little Endian) of the following ciphertext + tag |
| 4      | 12   | **Nonce** (Random IV for this chunk) |
| 16     | N    | **Ciphertext** (Encrypted Data + 16 byte Auth Tag) |

The file ends when the read loop encounters EOF or a 0-length chunk (if used as terminator, though current implementation relies on EOF).

## 3. Data Pipeline

### Encryption Flow
1. **Input**: Directory Traversal (recursive)
2. **Archive**: `tar` stream (Preserves metadata: permissions, timestamps, symlinks)
3. **Compression**: `zstd` (Level: SpeedBestCompression)
4. **Encryption**: Chunked AES-256-GCM
5. **Output**: File

### Decryption Flow
1. **Input**: File
2. **Decryption**: Chunked AES-256-GCM (Verify Auth Tag per chunk)
3. **Decompression**: `zstd`
4. **Extraction**: `tar` (Restores metadata)
5. **Output**: Directory

## 4. Dependencies

- `github.com/spf13/cobra`: CLI framework.
- `github.com/klauspost/compress`: Optimized Zstandard implementation.
- `golang.org/x/crypto/argon2`: Password hashing.
- `github.com/schollz/progressbar/v3`: Progress visualization.
