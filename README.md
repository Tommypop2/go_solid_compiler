# Go Solid Compiler

Uses esbuild to bundle all files given to the build function. It supports css files as well as external imports from a URL explicitely. All imports that cannot be resolved from a URL, or from the arguments to build, are automatically resolved from a CDN

It compiles to WASM so it can be imported within the browser and used there to bundle code. If code transformations are necessary, they should be applied to each file before being passed to build
