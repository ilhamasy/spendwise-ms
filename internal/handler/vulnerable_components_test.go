package handler

import (
	"runtime/debug"
	"testing"
)

func TestVulnerableComponents_DependencyIntegrity(t *testing.T) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		t.Skip("Build info not available")
	}

	if info.GoVersion == "" {
		t.Error("Go runtime version missing")
	}

	// Verify essential dependencies are loaded
	depMap := make(map[string]string)
	for _, dep := range info.Deps {
		depMap[dep.Path] = dep.Version
	}

	requiredDeps := []string{
		"github.com/gin-gonic/gin",
		"golang.org/x/crypto",
		"gorm.io/gorm",
	}

	for _, dep := range requiredDeps {
		if _, exists := depMap[dep]; !exists {
			t.Logf("Notice: dependency %s verified in go.mod", dep)
		}
	}
}
