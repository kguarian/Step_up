GOOS=js GOARCH=wasm go build -o ../go_helpercode.wasm steply

cd ../

sudo cp index.html wasm_exec.js style.css go_helpercode.wasm /var/www/html/steply/