package javascript

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ZacxDev/go-static-site/config"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/pkg/errors"
)

var isProd = os.Getenv("NODE_ENV")

// CompileJSTarget compiles JavaScript targets using either direct esbuild API or Node-based execution with polyfills
func CompileJSTarget(targets map[string]config.JavascriptTarget, usePolyfill bool, translations map[string]map[string]string, lang string) (map[string][]string, error) {
	if usePolyfill {
		res, err := compileWithNodePolyfill(targets, translations, lang)
		if err != nil {
			fmt.Printf("%+v\n", err)
			return nil, errors.WithStack(err)
		}

		return res, nil
	}
	return compileWithEsbuild(targets, translations, lang)
}

func marshalTranslations(translations map[string]map[string]string, lang string) string {
	b, _ := json.Marshal(translations[lang])
	return string(b)
}

// compileWithEsbuild uses the direct esbuild API (original implementation)
func compileWithEsbuild(targets map[string]config.JavascriptTarget, translations map[string]map[string]string, lang string) (map[string][]string, error) {
	emitted := make(map[string][]string, 0)
	for targetName, target := range targets {
		result := api.Build(api.BuildOptions{
			EntryPoints:       []string{target.Source},
			Bundle:            true,
			MinifyWhitespace:  true,
			MinifyIdentifiers: true,
			MinifySyntax:      true,
			TreeShaking:       api.TreeShakingTrue,
			Platform:          api.PlatformBrowser,
			Engines: []api.Engine{
				{Name: api.EngineChrome, Version: "100"},
				{Name: api.EngineFirefox, Version: "100"},
				{Name: api.EngineSafari, Version: "15"},
				{Name: api.EngineEdge, Version: "100"},
			},
			Loader: map[string]api.Loader{
				".css": api.LoaderText,
			},
			Sourcemap: api.SourceMapExternal,
			Write:     false,
			Outdir:    target.OutDir,
			Define: map[string]string{
				"__Lang__": fmt.Sprintf("%s", marshalTranslations(translations, lang)),
			},
		})

		if len(result.Errors) > 0 {
			return nil, errors.New(fmt.Sprintf("Esbuild error: %+v %+v", result.Errors[0], result.Errors[0].Location))
		}

		emittedPaths, err := processOutputFiles(result.OutputFiles, target, lang)
		if err != nil {
			return nil, err
		}
		emitted[targetName] = emittedPaths
	}
	return emitted, nil
}

// compileWithNodePolyfill uses Node to run esbuild with the node modules polyfill plugin
func compileWithNodePolyfill(targets map[string]config.JavascriptTarget, translations map[string]map[string]string, lang string) (map[string][]string, error) {
	emitted := make(map[string][]string, 0)

	// TODO: fix the extra .js file with no hash being emitted
	// Create temporary build script
	buildScript := `
const esbuild = require('esbuild');
const { polyfillNode } = require("esbuild-plugin-polyfill-node");

async function build(config) {
        try {
                const result = await esbuild.build({
                        ...config,
                        plugins: [polyfillNode()],
                        write: true,
                        metafile: true,
                });
                console.log(JSON.stringify(result.metafile));
        } catch (err) {
                console.log(JSON.stringify({ error: err.message }));
                process.exit(1);
        }
}

const config = JSON.parse(process.argv[2]);
build(config);
`

	scriptPath := "tmp-esbuild-script.js"
	defer os.Remove(scriptPath)
	if err := os.WriteFile(scriptPath, []byte(buildScript), 0644); err != nil {
		return nil, errors.WithStack(err)
	}

	for targetName, target := range targets {
		// Prepare esbuild configuration
		config := map[string]interface{}{
			"entryPoints":       []string{target.Source},
			"bundle":            true,
			"minifyWhitespace":  true,
			"minifyIdentifiers": true,
			"minifySyntax":      true,
			"treeShaking":       true,
			"platform":          "browser",
			"sourcemap":         true,
			"write":             false,
			"outdir":            target.OutDir,
			"target": []string{
				"chrome100",
				"firefox100",
				"safari15",
				"edge100",
			},
			"define": map[string]string{
				"Lang": fmt.Sprintf("%q", marshalTranslations(translations, lang)),
			},
		}

		configJSON, err := json.Marshal(config)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		// Execute Node script
		cmd := exec.Command("node", scriptPath, string(configJSON))
		output, err := cmd.Output()
		if err != nil {
			// Try to parse error from JSON output
			var errorResult struct {
				Error string `json:"error"`
			}
			if jsonErr := json.Unmarshal(output, &errorResult); jsonErr == nil && errorResult.Error != "" {
				return nil, errors.New("Esbuild error: " + errorResult.Error)
			}
			return nil, errors.WithStack(err)
		}

		// Parse metafile output
		var metafile struct {
			Outputs map[string]struct {
				Imports    []any  `json:"imports"`
				Exports    []any  `json:"exports"`
				EntryPoint string `json:"entryPoint"`
				Bytes      int    `json:"bytes"`
				Inputs     map[string]struct {
					BytesInOutput int `json:"bytesInOutput"`
				} `json:"inputs"`
			} `json:"outputs"`
		}
		if err := json.Unmarshal(output, &metafile); err != nil {
			return nil, errors.WithStack(err)
		}

		outFiles := make([]api.OutputFile, 0)
		for k := range metafile.Outputs {
			content, err := os.ReadFile(k)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			if k != "" {
				path, err := filepath.Abs(k)
				if err != nil {
					return nil, errors.WithStack(err)
				}

				hash, err := generateShortHash(k, 8)
				if err != nil {
					return nil, errors.WithStack(err)
				}

				outFiles = append(outFiles, api.OutputFile{
					Path:     path,
					Contents: content,
					Hash:     hash,
				})
			}
		}

		// Process outputs
		emittedPaths, err := processOutputFiles(outFiles, target, lang)
		if err != nil {
			return nil, err
		}
		emitted[targetName] = emittedPaths
	}

	return emitted, nil
}

// processOutputFiles handles the output files from esbuild (helper function for the original implementation)
func processOutputFiles(outputFiles []api.OutputFile, target config.JavascriptTarget, lang string) ([]string, error) {
	// Separate files with and without .map extension
	var regularFiles []api.OutputFile
	var mapFiles []api.OutputFile
	emitted := make([]string, 0)

	for _, out := range outputFiles {
		ext := filepath.Ext(out.Path)
		if strings.EqualFold(ext, ".map") {
			mapFiles = append(mapFiles, out)
		} else {
			regularFiles = append(regularFiles, out)
		}
	}

	// Concatenate regular files followed by .map files
	sortedFiles := append(regularFiles, mapFiles...)

	srcToHash := make(map[string]map[string]string)

	for _, out := range sortedFiles {
		// Modify the file path to include the hash
		dir := filepath.Dir(out.Path) // Get the directory of the original path
		dotJsIndex := strings.LastIndex(out.Path, ".js")
		dotCssIndex := strings.LastIndex(out.Path, ".css")

		var isMap bool
		var ext string
		var srcHashKey string
		if dotJsIndex != -1 {
			ext = out.Path[dotJsIndex:]
			isMap = ext == ".js.map"
			srcHashKey = "js"
		} else if dotCssIndex != -1 {
			ext = out.Path[dotCssIndex:]
			isMap = ext == ".css.map"
			srcHashKey = "css"
		} else {
			fmt.Printf("Warning: not emitting esbuild file due to unsupported file type: %s\n", out.Path)
			continue
		}

		base := filepath.Base(out.Path)                 // Get the file name with extension
		fileNameWithoutExt := base[:len(base)-len(ext)] // Get the file name without extension

		var hashForFileName string
		if isMap {
			hashForFileName = srcToHash[srcHashKey][fileNameWithoutExt]
			if hashForFileName == "" {
				msg := fmt.Sprintf("source map %s can not find hash for it's source file", fileNameWithoutExt)
				return nil, errors.New(msg)
			}
		} else {
			safeHash := strings.ReplaceAll(out.Hash, "/", "")
			if srcToHash[srcHashKey] != nil {
				srcToHash[srcHashKey][fileNameWithoutExt] = safeHash
			} else {
				srcToHash[srcHashKey] = map[string]string{
					fileNameWithoutExt: safeHash,
				}
			}
			hashForFileName = safeHash
		}

		// Create new path with hash included
		name := fmt.Sprintf("%s_%s_%s%s", lang, fileNameWithoutExt, hashForFileName, ext)
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
			emitted = append(emitted, publicPath)
		}
	}

	return emitted, nil
}

func generateShortHash(filePath string, length int) (string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create a SHA-256 hasher
	hasher := sha256.New()

	// Read and hash the file contents
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to hash file contents: %w", err)
	}

	// Get the full hash as a hex string
	fullHash := hex.EncodeToString(hasher.Sum(nil))

	// Truncate the hash to the desired length
	if length > len(fullHash) {
		length = len(fullHash) // Avoid out-of-range slicing
	}
	shortHash := fullHash[:length]

	return shortHash, nil
}
