# Adaclare Xtar - Secure Archiver

[![Go](https://github.com/quietvw/Xtar/actions/workflows/go.yml/badge.svg)](https://github.com/quietvw/Xtar/actions/workflows/go.yml)

Adaclare Xtar is a command-line tool for securely compressing and splitting `.tar.gz` files using AES-256 encryption. It also supports joining and decrypting split files for decompression. This tool is designed to allow efficient compression and encryption, providing the ability to split large files into smaller chunks for easier storage or transfer.

## Features

- **Compress and split**: Compress and split a `.tar.gz` file into smaller chunks, with optional AES-256 encryption.
- **Decompress and join**: Join and decrypt the chunks to recover the original file.
- **AES-256 encryption**: Encrypt and decrypt using a secure AES-256 encryption algorithm.
- **Progress indicator**: Visual feedback on compression and decompression progress with a percentage and spinner.
- **File splitting**: Split large files based on the size limit you specify (e.g., 100MB per chunk).

## Requirements

- Go 1.18 or higher
- AES-256 encryption library (included in the Go standard library)

## Installation

To install `Adaclare Xtar`, clone this repository and build the binary using Go:

```bash
git clone https://github.com/your-username/adaclare-xtar.git
cd adaclare-xtar
go build -o adaclare-xtar

