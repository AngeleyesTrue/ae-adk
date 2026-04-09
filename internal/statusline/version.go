package statusline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/AngeleyesTrue/ae-adk/internal/defs"
	"gopkg.in/yaml.v3"
)

// VersionCollector는 설정 파일에서 AE 템플릿 버전을 읽고
// 실행 중인 바이너리 버전과 비교하여 `ae update`를 통한
// 템플릿 동기화 필요 여부를 판단한다.
type VersionCollector struct {
	mu            sync.RWMutex
	cached        string
	binaryVersion string // 실행 중인 바이너리 버전 (pkg/version에서 전달)
}

// VersionConfig는 .ae/config/sections/system.yaml의 버전 필드 파싱 구조체이다.
// ae.template_version(ae update로 갱신)을 우선 읽고,
// 없으면 ae.version(초기화 시 설정)으로 폴백한다.
type VersionConfig struct {
	AE struct {
		Version         string `yaml:"version"`
		TemplateVersion string `yaml:"template_version"`
	} `yaml:"ae"`
}

// effectiveVersion은 ae.template_version이 설정되어 있으면 반환하고,
// 없으면 ae.version을 반환한다. 이를 통해 statusline이 초기화 버전이 아닌
// `ae update`로 갱신된 버전을 반영하도록 보장한다.
func (c *VersionConfig) effectiveVersion() string {
	if c.AE.TemplateVersion != "" {
		return c.AE.TemplateVersion
	}
	return c.AE.Version
}

// NewVersionCollector는 .ae/config/sections/system.yaml에서 템플릿 버전을
// 읽어 실행 중인 바이너리 버전과 비교하는 VersionCollector를 생성한다.
// binaryVersion이 비어있으면 템플릿 버전을 폴백으로 사용한다.
func NewVersionCollector(binaryVersion string) *VersionCollector {
	return &VersionCollector{binaryVersion: binaryVersion}
}

// CheckUpdate는 설정 파일에서 템플릿 버전을 읽고 바이너리 버전과 비교한다.
// Current에는 바이너리 버전이 설정되며, 바이너리와 템플릿 버전이 다르면
// SyncNeeded=true와 TemplateVersion이 설정된다.
// 설정 파일이 없거나 버전이 없으면 바이너리 버전만 표시한다.
func (v *VersionCollector) CheckUpdate(_ context.Context) (*VersionData, error) {
	// 캐시 확인
	v.mu.RLock()
	if v.cached != "" {
		version := v.cached
		v.mu.RUnlock()
		return v.buildVersionData(version), nil
	}
	v.mu.RUnlock()

	// 설정 파일 검색 및 읽기
	version, err := v.readVersionFromConfig()
	if err != nil {
		// 폴백: 설정 파일이 없으면 바이너리 버전 사용
		if v.binaryVersion != "" {
			return &VersionData{
				Current:   formatVersion(v.binaryVersion),
				Available: true,
			}, nil
		}
		return &VersionData{Available: false}, nil
	}

	// 캐시 갱신
	v.mu.Lock()
	v.cached = version
	v.mu.Unlock()

	return v.buildVersionData(version), nil
}

// buildVersionData는 바이너리 버전을 Current로 설정하고,
// 템플릿 버전과 다를 경우 SyncNeeded를 true로 표시한다.
// Latest/UpdateAvailable은 향후 원격 버전 체크를 위해 예약되어 있다.
func (v *VersionCollector) buildVersionData(templateVersion string) *VersionData {
	bv := formatVersion(v.binaryVersion)
	tv := formatVersion(templateVersion)

	// Current: 바이너리 버전 우선, 없으면 템플릿 버전으로 폴백
	current := bv
	if current == "" {
		current = tv
	}

	data := &VersionData{
		Current:   current,
		Available: true,
	}

	// 바이너리와 템플릿 버전이 모두 존재하고 다르면 동기화 필요
	if bv != "" && tv != "" && bv != tv {
		data.SyncNeeded = true
		data.TemplateVersion = tv
	}

	return data
}

// readVersionFromConfig는 현재 디렉토리에서 상위로 올라가며
// .ae/config/sections/system.yaml을 찾아 프로젝트 루트를 탐색한다.
func (v *VersionCollector) readVersionFromConfig() (string, error) {
	// 현재 디렉토리에서 시작
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	// .ae/config/sections/system.yaml을 상위로 탐색
	for {
		configPath := filepath.Join(dir, defs.AEDir, defs.SectionsSubdir, defs.SystemYAML)
		if _, err := os.Stat(configPath); err == nil {
			// 설정 파일 발견
			return v.parseConfigFile(configPath)
		}

		// 상위 디렉토리로 이동
		parent := filepath.Dir(dir)
		if parent == dir {
			// 파일시스템 루트에 도달
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("config file not found")
}

// parseConfigFile은 설정 파일을 읽고 파싱하여 버전을 추출한다.
func (v *VersionCollector) parseConfigFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read config file: %w", err)
	}

	var cfg VersionConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("parse config file: %w", err)
	}

	version := cfg.effectiveVersion()
	if version == "" {
		return "", fmt.Errorf("version not set in config")
	}

	return version, nil
}

// formatVersion은 버전 문자열에서 'v' 접두사를 제거한다.
// 예: "v1.14.0" -> "1.14.0", "1.14.0" -> "1.14.0"
func formatVersion(v string) string {
	return strings.TrimPrefix(v, "v")
}
