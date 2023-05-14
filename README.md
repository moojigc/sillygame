# sillygame
Multiplayer drawings! Using the power of websockets, draw collaboratively with your friends!

# Build from source
Prereqs:
- `go` toolchain
- `node` and `npm`

From root directory, run:

    export VITE_WEBSOCKET_URL=localhost:3000/subscribe
    export VITE_WEB_URL=http://localhost:3000
    go build -o game ./cmd/server
    cd web && npm run build && mv build ../static && cd ..
Now you can run it:

    ./game localhost:3000
