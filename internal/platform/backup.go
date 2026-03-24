package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	backupTimeFormat = "20060102-150405"
	maxBackups       = 5
)

// BackupSettings는 settings.json을 타임스탬프 기반으로 백업한다.
func BackupSettings(sys SystemInfo, settingsPath string) (string, error) {
	if !sys.FileExists(settingsPath) {
		return "", fmt.Errorf("settings file not found: %s", settingsPath)
	}

	data, err := sys.ReadFile(settingsPath)
	if err != nil {
		return "", fmt.Errorf("read settings: %w", err)
	}

	timestamp := time.Now().Format(backupTimeFormat)
	backupPath := settingsPath + "." + timestamp + ".bak"

	if err := sys.WriteFile(backupPath, data, 0o644); err != nil {
		return "", fmt.Errorf("write backup: %w", err)
	}

	return backupPath, nil
}

// CleanupOldBackups는 오래된 백업 파일을 삭제하여 최근 maxBackups개만 유지한다.
func CleanupOldBackups(settingsPath string) error {
	dir := filepath.Dir(settingsPath)
	base := filepath.Base(settingsPath)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read backup directory: %w", err)
	}

	var backups []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, base+".") && strings.HasSuffix(name, ".bak") {
			backups = append(backups, filepath.Join(dir, name))
		}
	}

	if len(backups) <= maxBackups {
		return nil
	}

	// os.ReadDir은 이미 이름순 정렬을 반환하므로 추가 정렬 불필요
	// 오래된 백업 삭제
	toDelete := backups[:len(backups)-maxBackups]
	for _, path := range toDelete {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove old backup %s: %w", path, err)
		}
	}

	return nil
}
