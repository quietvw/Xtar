package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	appVersion = "1.0.0"
	chunkFmt   = ".%02d"
)

func main() {
	showBanner()

	// Flags
	compressFile := flag.String("c", "", "Compress (split) an existing .tar.gz file")
	decompressFile := flag.String("d", "", "Decompress a .tar.gz split archive")
	encryptionKey := flag.String("e", "", "AES-256 encryption key (hex string)")
	splitSizeStr := flag.String("s", "", "Split size limit (e.g., 100M)")
	version := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *version {
		fmt.Printf("Adaclare Xtar ( https://github.com/quietvw/Xtar ) version %s\n", appVersion)
		return
	}

	if *compressFile == "" && *decompressFile == "" {
		fmt.Println("Error: Use -c to compress or -d to decompress.")
		flag.Usage()
		os.Exit(1)
	}

	var key []byte
	var err error
	if *encryptionKey != "" {
		key, err = hex.DecodeString(*encryptionKey)
		if err != nil || (len(key) != 16 && len(key) != 24 && len(key) != 32) {
			fmt.Println("Encryption key must be valid hex (16/24/32 bytes).")
			os.Exit(1)
		}
	}

	if *compressFile != "" {
		splitSize := parseSize(*splitSizeStr)
		err = compressAndSplit(*compressFile, splitSize, key)
	} else if *decompressFile != "" {
		err = joinAndDecompress(*decompressFile, key)
	}

	if err != nil {
		fmt.Println("❌ Failed:", err)
		os.Exit(1)
	}
	fmt.Println("✅ Done.")
}

func showBanner() {
	fmt.Println(`========================================`)
	fmt.Println(`    🔐 Adaclare Xtar - Secure Archiver`)
	fmt.Println(`         https://adaclare.com`)
	fmt.Println(`     (c) 2025 Adaclare Corporation`)
	fmt.Println(`========================================`)
}

func parseSize(sizeStr string) int64 {
	if sizeStr == "" {
		return 0
	}
	mult := int64(1)
	switch {
	case strings.HasSuffix(sizeStr, "K"):
		mult = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "K")
	case strings.HasSuffix(sizeStr, "M"):
		mult = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "M")
	case strings.HasSuffix(sizeStr, "G"):
		mult = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "G")
	}
	val, _ := strconv.ParseInt(sizeStr, 10, 64)
	return val * mult
}

func compressAndSplit(file string, splitSize int64, key []byte) error {
	input, err := os.Open(file)
	if err != nil {
		return err
	}
	defer input.Close()

	// Get total size of the file for progress calculation
	stat, err := input.Stat()
	if err != nil {
		return err
	}
	totalSize := stat.Size()

	fmt.Printf("🔧 Compressing %s...\n", file)
	part := 0
	var current *os.File
	var writer io.Writer
	buf := make([]byte, 4096)
	var written int64

	done := make(chan bool)
	showProgress(done, "Compressing", totalSize, &written)

	writeChunk := func() error {
		if current != nil {
			current.Close()
		}
		name := fmt.Sprintf("%s"+chunkFmt, file, part)
		current, err = os.Create(name)
		if err != nil {
			return err
		}
		writer = current
		if key != nil {
			writer, err = encryptWriter(current, key)
		}
		part++
		return err
	}

	err = writeChunk()
	if err != nil {
		return err
	}

	for {
		n, err := input.Read(buf)
		if n > 0 {
			if splitSize > 0 && written+int64(n) > splitSize {
				excess := splitSize - written
				writer.Write(buf[:excess])
				writeChunk()
				writer.Write(buf[excess:n])
				written = int64(n) - excess
			} else {
				writer.Write(buf[:n])
				written += int64(n)
			}
		}
		if err == io.EOF {
			break
		}
	}
	if current != nil {
		current.Close()
	}
	done <- true
	return nil
}

func joinAndDecompress(base string, key []byte) error {
	outFileName := "joined_" + filepath.Base(base)
	outFile, err := os.Create(outFileName)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Get the total size of all the parts for progress calculation
	var totalSize int64
	for i := 0; ; i++ {
		part := fmt.Sprintf("%s"+chunkFmt, base, i)
		if _, err := os.Stat(part); errors.Is(err, os.ErrNotExist) {
			break
		}
		partStat, err := os.Stat(part)
		if err != nil {
			return err
		}
		totalSize += partStat.Size()
	}

	fmt.Printf("📦 Decompressing parts to %s\n", outFileName)
	done := make(chan bool)
	var written int64
	showProgress(done, "Decompressing", totalSize, &written)

	for i := 0; ; i++ {
		part := fmt.Sprintf("%s"+chunkFmt, base, i)
		if _, err := os.Stat(part); errors.Is(err, os.ErrNotExist) {
			break
		}
		in, err := os.Open(part)
		if err != nil {
			return err
		}
		var reader io.Reader = in
		if key != nil {
			reader, err = decryptReader(in, key)
			if err != nil {
				return fmt.Errorf("error decrypting %s: %v", part, err)
			}
		}
		partStat, err := in.Stat()
		if err != nil {
			return err
		}
		_, err = io.Copy(outFile, reader)
		in.Close()
		if err != nil {
			return err
		}
		written += partStat.Size()
	}
	done <- true
	return nil
}

func showProgress(done <-chan bool, msg string, totalSize int64, currentSize *int64) {
	go func() {
		spinner := []rune{'|', '/', '-', '\\'}
		i := 0
		for {
			select {
			case <-done:
				fmt.Printf("\r%s... done\n", msg)
				return
			default:
				progress := float64(*currentSize) / float64(totalSize) * 100
				fmt.Printf("\r%s... %c %.2f%%", msg, spinner[i%len(spinner)], progress)
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func encryptWriter(w io.Writer, key []byte) (io.Writer, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, aes.BlockSize)
	rand.Read(iv)
	w.Write(iv)
	w.Write(checksum(key)[:4])
	stream := cipher.NewCFBEncrypter(block, iv)
	return &cipher.StreamWriter{S: stream, W: w}, nil
}

func decryptReader(r io.Reader, key []byte) (io.Reader, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(r, iv); err != nil {
		return nil, err
	}
	check := make([]byte, 4)
	if _, err := io.ReadFull(r, check); err != nil {
		return nil, err
	}
	if !bytes.Equal(check, checksum(key)[:4]) {
		return nil, errors.New("encryption key mismatch or tampered data")
	}
	stream := cipher.NewCFBDecrypter(block, iv)
	return &cipher.StreamReader{S: stream, R: r}, nil
}

func checksum(key []byte) []byte {
	h := sha256.Sum256(key)
	return h[:]
}
