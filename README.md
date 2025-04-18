# Adaclare Xtar - Secure Archiver

[![Go](https://github.com/quietvw/Xtar/actions/workflows/go.yml/badge.svg)](https://github.com/quietvw/Xtar/actions/workflows/go.yml)

Adaclare Xtar is a command-line tool for securely compressing and splitting `.tar.gz` files using AES-256 encryption. It also supports joining and decrypting split files for decompression. This tool is designed to allow efficient compression and encryption, providing the ability to split large files into smaller chunks for easier storage or transfer.

---

## Features

- **Compress and split**: Compress and split a `.tar.gz` file into smaller chunks, with optional AES-256 encryption.
- **Decompress and join**: Join and decrypt the chunks to recover the original file.
- **AES-256 encryption**: Encrypt and decrypt using a secure AES-256 encryption algorithm.
- **Progress indicator**: Visual feedback on compression and decompression progress with a percentage and spinner.
- **File splitting**: Split large files based on the size limit you specify (e.g., 100MB per chunk).

---

## Requirements

- Go 1.18 or higher
- Uses standard Go libraries (`crypto/aes`, `crypto/cipher`, etc.)

---

## Installation

```bash
git clone https://github.com/quietvw/Xtar.git
cd Xtar
go build -o xtar xtar.go
```

---

## 🔧 Sample Usage

### 1. Generate an Encryption Key

Generate a random 128-bit (16-byte) AES key:

```bash
openssl rand -hex 16
# Example output: 3f9c2b7e8d5a1046b1e3c7f2d98a6721
```

You can also use a 192-bit or 256-bit key (24 or 32 bytes in hex) if desired.

---

### 2. Compress and Split with Optional Encryption

```bash
./xtar -c mydata.tar.gz -s 50M -e 3f9c2b7e8d5a1046b1e3c7f2d98a6721
```

- `-c`: Path to your `.tar.gz` file
- `-s`: Maximum split size (e.g. `50M`)
- `-e`: (Optional) AES encryption key in hex (must be 16, 24, or 32 bytes)

Output:

```bash
🔧 Compressing mydata.tar.gz...
Compressing... | 42.13%
Compressing... / 78.99%
Compressing... - 100.00%
✅ Done.
```

Split output files:
```
mydata.tar.gz.00
mydata.tar.gz.01
mydata.tar.gz.02
...
```

---

### 3. Join and Decrypt

To reassemble the original file from chunks:

```bash
./xtar -d mydata.tar.gz -e 3f9c2b7e8d5a1046b1e3c7f2d98a6721
```

- `-d`: Base name of split files (omit the `.00`, `.01`, etc.)
- `-e`: Same key used during encryption

Output:

```bash
📦 Decompressing parts to joined_mydata.tar.gz
Decompressing... | 50.00%
Decompressing... / 85.00%
Decompressing... - 100.00%
✅ Done.
```

Final output:
```
joined_mydata.tar.gz
```

---

### 4. Show Version

```bash
./xtar -version
# Adaclare Xtar version 1.0.0
```

---

## 🔒 Notes

- Use strong, secure keys (16/24/32 bytes hex format).
- Do **not** lose your encryption key — files encrypted with it cannot be recovered without it.
- Only `.tar.gz` files are supported for now.
- Ideal for secure cloud backup workflows, archiving logs, or sensitive data.

---

## License

This project is licensed under the GNU General Public License. See the [LICENSE](LICENSE) file for details.

---

## Developed by

**Adaclare Corporation**  
[https://adaclare.com](https://adaclare.com)
