package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ArtifactRegistry tracks generated files for cleanup purposes
type ArtifactRegistry struct {
	BuildTimestamp time.Time         `json:"build_timestamp"`
	GeneratedFiles []string          `json:"generated_files"`
	StaticFiles    []string          `json:"static_files"`
	AssetFiles     []string          `json:"asset_files"`
	OutputDir      string            `json:"output_dir"`
	ManifestPath   string            `json:"manifest_path"`
	Metadata       map[string]string `json:"metadata"`
}

const artifactRegistryFile = ".build-artifacts.json"

// LoadArtifactRegistry loads the artifact registry from disk
func LoadArtifactRegistry(outputDir string) (*ArtifactRegistry, error) {
	registryPath := filepath.Join(outputDir, artifactRegistryFile)

	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		return &ArtifactRegistry{
			GeneratedFiles: []string{},
			StaticFiles:    []string{},
			AssetFiles:     []string{},
			OutputDir:      outputDir,
			Metadata:       make(map[string]string),
		}, nil
	}

	data, err := os.ReadFile(registryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read artifact registry: %w", err)
	}

	var registry ArtifactRegistry
	err = json.Unmarshal(data, &registry)
	if err != nil {
		return nil, fmt.Errorf("failed to parse artifact registry: %w", err)
	}

	return &registry, nil
}

// SaveArtifactRegistry saves the artifact registry to disk
func (ar *ArtifactRegistry) SaveArtifactRegistry() error {
	ar.BuildTimestamp = time.Now()

	registryPath := filepath.Join(ar.OutputDir, artifactRegistryFile)
	data, err := json.MarshalIndent(ar, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal artifact registry: %w", err)
	}

	err = os.WriteFile(registryPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write artifact registry: %w", err)
	}

	return nil
}

// AddGeneratedFile adds a generated file to the registry
func (ar *ArtifactRegistry) AddGeneratedFile(filePath string) {
	// Make path relative to output directory
	relPath, err := filepath.Rel(ar.OutputDir, filePath)
	if err != nil {
		relPath = filePath
	}
	ar.GeneratedFiles = append(ar.GeneratedFiles, relPath)
}

// AddStaticFile adds a static file to the registry
func (ar *ArtifactRegistry) AddStaticFile(filePath string) {
	// Make path relative to output directory
	relPath, err := filepath.Rel(ar.OutputDir, filePath)
	if err != nil {
		relPath = filePath
	}
	ar.StaticFiles = append(ar.StaticFiles, relPath)
}

// AddAssetFile adds a JS/CSS asset file to the registry
func (ar *ArtifactRegistry) AddAssetFile(filePath string) {
	// Make path relative to output directory
	relPath, err := filepath.Rel(ar.OutputDir, filePath)
	if err != nil {
		relPath = filePath
	}
	ar.AssetFiles = append(ar.AssetFiles, relPath)
}

// CleanupOrphanedFiles removes files that were generated in previous builds
// but are no longer needed in the current build
func (ar *ArtifactRegistry) CleanupOrphanedFiles(newRegistry *ArtifactRegistry) error {
	var removedFiles []string

	// Create maps for quick lookup of new files
	newGenerated := make(map[string]bool)
	newStatic := make(map[string]bool)
	newAssets := make(map[string]bool)

	for _, file := range newRegistry.GeneratedFiles {
		newGenerated[file] = true
	}
	for _, file := range newRegistry.StaticFiles {
		newStatic[file] = true
	}
	for _, file := range newRegistry.AssetFiles {
		newAssets[file] = true
	}

	// Remove orphaned generated files
	for _, file := range ar.GeneratedFiles {
		if !newGenerated[file] {
			fullPath := filepath.Join(ar.OutputDir, file)
			if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
				fmt.Printf("Warning: failed to remove orphaned file %s: %v\n", fullPath, err)
			} else if err == nil {
				removedFiles = append(removedFiles, file)
				fmt.Printf("Removed orphaned generated file: %s\n", file)
			}
		}
	}

	// Remove orphaned static files
	for _, file := range ar.StaticFiles {
		if !newStatic[file] {
			fullPath := filepath.Join(ar.OutputDir, file)
			if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
				fmt.Printf("Warning: failed to remove orphaned file %s: %v\n", fullPath, err)
			} else if err == nil {
				removedFiles = append(removedFiles, file)
				fmt.Printf("Removed orphaned static file: %s\n", file)
			}
		}
	}

	// Remove orphaned asset files
	for _, file := range ar.AssetFiles {
		if !newAssets[file] {
			fullPath := filepath.Join(ar.OutputDir, file)
			if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
				fmt.Printf("Warning: failed to remove orphaned file %s: %v\n", fullPath, err)
			} else if err == nil {
				removedFiles = append(removedFiles, file)
				fmt.Printf("Removed orphaned asset file: %s\n", file)
			}
		}
	}

	if len(removedFiles) > 0 {
		fmt.Printf("Cleaned up %d orphaned files\n", len(removedFiles))
	}

	return nil
}

// RemoveEmptyDirectories removes empty directories from the output directory
func (ar *ArtifactRegistry) RemoveEmptyDirectories() error {
	return filepath.Walk(ar.OutputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && path != ar.OutputDir {
			// Check if directory is empty
			entries, err := os.ReadDir(path)
			if err != nil {
				return err
			}

			if len(entries) == 0 {
				fmt.Printf("Removing empty directory: %s\n", path)
				return os.Remove(path)
			}
		}

		return nil
	})
}

// GetBuildStats returns statistics about the current build
func (ar *ArtifactRegistry) GetBuildStats() map[string]interface{} {
	return map[string]interface{}{
		"build_timestamp":   ar.BuildTimestamp,
		"generated_files":   len(ar.GeneratedFiles),
		"static_files":      len(ar.StaticFiles),
		"asset_files":       len(ar.AssetFiles),
		"total_files":       len(ar.GeneratedFiles) + len(ar.StaticFiles) + len(ar.AssetFiles),
		"output_directory":  ar.OutputDir,
		"manifest_path":     ar.ManifestPath,
	}
}