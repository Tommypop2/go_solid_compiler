package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"syscall/js"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/lithammer/dedent"
)

func CDN_URL(importee string) string {
	return "https://jspm.dev/" + importee
}

func uid(str string) int {
	s := 0
	strBytes := []byte(str)
	for i := 0; i < len(strBytes); i++ {
		c := int(str[i])
		s = ((31 * s) + c)
	}
	return s
}

var importMap map[string]interface{}
var fileMap map[string]string
var EntryPoint string = "./main"
var replPlugin = api.Plugin{
	Name: "replPlugin",
	Setup: func(build api.PluginBuild) {
		importMap = make(map[string]interface{})
		build.OnResolve(api.OnResolveOptions{Filter: ".*"}, func(args api.OnResolveArgs) (api.OnResolveResult, error) {
			importee := args.Path
			if strings.HasPrefix(importee, ".") {
				return api.OnResolveResult{Path: importee, Namespace: "replPlugin"}, nil
			}
			if strings.Contains(importee, "://") {
				return api.OnResolveResult{Path: importee, Namespace: "replPlugin", External: true}, nil
			}
			cdn_url := CDN_URL(importee)
			importMap[importee] = cdn_url
			return api.OnResolveResult{Path: importee, Namespace: "replPlugin", External: true}, nil
		})
		build.OnLoad(api.OnLoadOptions{Filter: ".*"}, func(args api.OnLoadArgs) (api.OnLoadResult, error) {
			filename := args.Path
			data := fileMap[filename]
			var contents string
			if strings.HasSuffix(filename, ".css") {
				contents = `
				(() => {
				  let stylesheet = document.getElementById('{id}');
				  if (!stylesheet) {
					stylesheet = document.createElement('style')
					stylesheet.setAttribute('id', {id})
					document.head.appendChild(stylesheet)
				  }
				  const styles = document.createTextNode({coolCode})
				  stylesheet.innerHTML = ''
				  stylesheet.appendChild(styles)
				})()
			  `
				contents = strings.ReplaceAll(contents, "{coolCode}", "`"+data+"`")
				contents = strings.ReplaceAll(contents, "{id}", strconv.Itoa(uid(filename)))
				contents = dedent.Dedent(contents)
			} else {
				contents = string(data)
			}
			return api.OnLoadResult{Contents: &contents}, nil
		})
	},
}

type JSData struct {
	Files [][]string
}

func build() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		var data JSData
		json.Unmarshal([]byte(args[0].String()), &data)
		files := data.Files
		fileMap = make(map[string]string)
		for i := 0; i < len(files); i++ {
			filename := files[i][0]
			fileContents := files[i][1]
			fileMap[filename] = fileContents
		}
		result := api.Build(api.BuildOptions{
			EntryPoints: []string{EntryPoint},
			Bundle:      true,
			Write:       false,
			Platform:    api.PlatformBrowser,
			LogLevel:    api.LogLevelInfo,
			Plugins:     []api.Plugin{replPlugin},
			Format:      api.FormatESModule,
		})
		file := result.OutputFiles[0].Contents
		return string(file)
	})
}
func getImportMap() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		return js.ValueOf(importMap)
	})
}
func setEntryPoint() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		newEntryPoint := args[0].String()
		EntryPoint = newEntryPoint
		return true
	})
}
func main() {
	js.Global().Set("build", build())
	js.Global().Set("getImportMap", getImportMap())
	js.Global().Set("setEntryPoint", setEntryPoint())
	js.Global().Call("onWasmLoaded")
	<-make(chan bool)
}
