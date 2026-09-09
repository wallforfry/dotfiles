package telemetry

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type CacheEntry struct {
	Digest  string  `json:"digest"`
	Source  string  `json:"source"`
	Summary Summary `json:"summary"`
}

type cacheFile struct {
	Version int                   `json:"version"`
	Since   string                `json:"since"`
	Files   map[string]CacheEntry `json:"files"`
}

func FileDigest(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func TranscriptFiles(root string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".jsonl") {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}

func LoadCache(path, since string) map[string]CacheEntry {
	content, err := os.ReadFile(path)
	if err != nil {
		return map[string]CacheEntry{}
	}
	var cache cacheFile
	if json.Unmarshal(content, &cache) != nil || cache.Version != CacheVersion || cache.Since != since || cache.Files == nil {
		return map[string]CacheEntry{}
	}
	return cache.Files
}

func SaveCache(path, since string, entries map[string]CacheEntry) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(directory, "telemetry.*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	encoder := json.NewEncoder(temporary)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(cacheFile{Version: CacheVersion, Since: since, Files: entries}); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}

func CollectSource(source, root, since string, cache map[string]CacheEntry) (map[string]CacheEntry, []Summary, int, error) {
	paths, err := TranscriptFiles(root)
	if err != nil {
		return nil, nil, 0, readInterrupted(source, err)
	}
	entries := make(map[string]CacheEntry, len(paths))
	summaries := make([]Summary, 0, len(paths))
	hits := 0
	for _, path := range paths {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return nil, nil, 0, readInterrupted(source, err)
		}
		identityBytes := sha256.Sum256([]byte(absolute))
		identity := hex.EncodeToString(identityBytes[:])
		digest, err := FileDigest(path)
		if err != nil {
			return nil, nil, 0, readInterrupted(source, err)
		}
		entry, found := cache[identity]
		var summary Summary
		if found && entry.Digest == digest && entry.Source == source && validCachedSummary(entry.Summary) {
			summary = entry.Summary
			hits++
		} else {
			summary, err = SummarizeFile(source, path, since)
			if err != nil {
				return nil, nil, 0, readInterrupted(source, err)
			}
		}
		entries[identity] = CacheEntry{Digest: digest, Source: source, Summary: summary}
		summaries = append(summaries, summary)
	}
	return entries, summaries, hits, nil
}

func validCachedSummary(summary Summary) bool {
	if summary.Source == "" || summary.Days == nil {
		return false
	}
	for _, field := range summaryFields {
		if field(&summary) == nil {
			return false
		}
	}
	return true
}
