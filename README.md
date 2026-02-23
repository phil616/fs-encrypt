# fs-encrypt - 目录加密与压缩工具

`fs-encrypt` 是一个企业级的命令行工具，用于安全地加密和压缩目录。它专为高安全性及高压缩率需求设计，支持多层级目录结构，并保留完整的文件系统元信息。

## 功能特性

1.  **全目录加密**：递归处理多层文件夹和文件，对文件名和文件内容进行全量加密。
2.  **高压缩率**：采用 Zstandard (zstd) 算法的最高压缩级别，尽可能节省存储空间。
3.  **高安全性**：
    *   **加密算法**：使用 AES-256-GCM 进行认证加密，确保数据的机密性和完整性。
    *   **密钥派生**：使用 Argon2id 算法从密码派生密钥，有效抵抗彩虹表和暴力破解攻击。
    *   **随机化**：每个加密块使用唯一的随机 Nonce，每个文件使用唯一的随机 Salt。
4.  **元数据保留**：完整保留文件的权限、修改时间等元信息（通过 Tar 归档格式）。
5.  **流式处理**：采用流式加密和解密，支持处理超大文件，内存占用低。
6.  **交互式 CLI**：友好的命令行交互，包含进度条提示和密码安全输入。

## 技术架构

*   **归档层**：使用 `tar` 格式将目录结构序列化，保留文件元数据。
*   **压缩层**：使用 `zstd` (Zstandard) 进行流式压缩，配置为 `SpeedBestCompression` 以获得最佳压缩比。
*   **加密层**：
    *   将数据流分割为 64KB 的块。
    *   每个块独立使用 AES-256-GCM 加密。
    *   文件头包含 16 字节随机 Salt。
    *   每个块包含 4 字节长度、12 字节随机 Nonce 和加密数据（含 Tag）。
*   **开发语言**：Go (Golang)。

## 编译指南

### 前置要求

*   Go 1.20 或更高版本。

### 编译步骤

1.  克隆项目：
    ```bash
    git clone https://github.com/your-repo/fs-encrypt.git
    cd fs-encrypt
    ```

2.  下载依赖：
    ```bash
    go mod tidy
    ```

3.  编译二进制文件：
    ```bash
    # Windows
    go build -o fs-encrypt.exe ./cmd/fs-encrypt

    # Linux / macOS
    go build -o fs-encrypt ./cmd/fs-encrypt
    ```

## 使用说明

### 加密目录

将 `target_dir` 目录加密并保存为 `output.enc`：

```bash
./fs-encrypt encrypt target_dir output.enc
```

程序会提示输入密码并确认。

### 解密文件

将 `output.enc` 解密并还原到 `restore_dir` 目录：

```bash
./fs-encrypt decrypt output.enc restore_dir
```

程序会提示输入解密密码。

## 最佳实践

*   **密码强度**：建议使用强密码，因为安全性完全依赖于密码强度。
*   **备份**：请妥善保管密码，一旦遗忘，数据将无法恢复（采用工业级加密，无后门）。
*   **性能**：由于使用了最高级别的压缩，加密过程可能较慢，请耐心等待。

## 目录结构

```
fs-encrypt/
├── cmd/
│   └── fs-encrypt/    # 主程序入口及命令定义
├── internal/
│   ├── archive/       # 归档与解包逻辑 (Tar)
│   ├── crypto/        # 加密与解密核心逻辑 (AES-GCM, Argon2)
│   └── utils/         # 通用工具
├── docs/              # 文档
└── go.mod             # 依赖定义
```

## 许可证

MIT License
