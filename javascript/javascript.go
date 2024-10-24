package javascript

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZacxDev/go-static-site/config"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/pkg/errors"
)

var isProd = os.Getenv("NODE_ENV")

func CompileJSTarget(targets map[string]config.JavascriptTarget) (map[string]string, error) {
	emitted := make(map[string]string, 0)
	for targetName, target := range targets {
		result := api.Build(api.BuildOptions{
			EntryPoints:       []string{target.Source},
			Bundle:            true,
			MinifyWhitespace:  true,
			MinifyIdentifiers: true,
			MinifySyntax:      true,
			Engines: []api.Engine{
				{Name: api.EngineChrome, Version: "100"},
				{Name: api.EngineFirefox, Version: "100"},
				{Name: api.EngineSafari, Version: "15"},
				{Name: api.EngineEdge, Version: "100"},
			},
			Sourcemap: api.SourceMapExternal,
			Write:     false,
			Outdir:    target.OutDir,
		})

		if len(result.Errors) > 0 {
			os.Exit(1)
		}

		// Separate files with and without .map extension
		var regularFiles []api.OutputFile
		var mapFiles []api.OutputFile

		for _, out := range result.OutputFiles {
			ext := filepath.Ext(out.Path)
			if strings.EqualFold(ext, ".map") {
				mapFiles = append(mapFiles, out)
			} else {
				regularFiles = append(regularFiles, out)
			}
		}

		// Concatenate regular files followed by .map files
		sortedFiles := append(regularFiles, mapFiles...)

		srcToHash := make(map[string]string)

		for _, out := range sortedFiles {
			// Modify the file path to include the hash
			dir := filepath.Dir(out.Path) // Get the directory of the original path
			ext := out.Path[strings.Index(out.Path, "."):]
			isMap := ext == ".js.map"
			base := filepath.Base(out.Path)                 // Get the file name with extension
			fileNameWithoutExt := base[:len(base)-len(ext)] // Get the file name without extension

			var hashForFileName string
			if isMap {
				hashForFileName = srcToHash[fileNameWithoutExt]
				if hashForFileName == "" {
					msg := fmt.Sprintf("source map %s can not find hash for it's source file", fileNameWithoutExt)
					return nil, errors.New(msg)
				}
			} else {
				safeHash := strings.ReplaceAll(out.Hash, "/", "")
				srcToHash[fileNameWithoutExt] = safeHash
				hashForFileName = safeHash
			}

			// Create new path with hash included
			name := fmt.Sprintf("%s_%s%s", fileNameWithoutExt, hashForFileName, ext)
			newPath := filepath.Join(dir, name)

			// Open the file, create if it doesn't exist, truncate if it does
			file, err := os.OpenFile(newPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				fmt.Printf("failed to open file %s: %s", newPath, err.Error())
				return nil, errors.WithStack(err)
			}

			var fileContentB []byte
			if isMap {
				fileContentB = out.Contents
			} else {
				// Write the contents to the file
				srcMap := fmt.Sprintf("//# sourceMappingURL=%s.map", name)
				fileContent := string(out.Contents) + srcMap
				fileContentB = []byte(fileContent)
			}

			_, err = file.Write(fileContentB)
			if err != nil {
				file.Close() // Ensure we close the file in case of an error
				fmt.Printf("failed to write to file %s: %s", newPath, err.Error())
				return nil, errors.WithStack(err)
			}

			// Close the file after writing
			err = file.Close()
			if err != nil {
				fmt.Printf("failed to close file %s: %s", newPath, err.Error())
				return nil, errors.WithStack(err)
			}

			if !isMap {
				publicPath := "/" + target.OutDir + "/" + name
				emitted[targetName] = publicPath
			}
		}
	}

	return emitted, nil
}
