# Maelstrom

```sh
oapi-codegen -generate chi-server -package maelstrom http/maelstrom.yaml > cmd/maelstrom/internal/maelstrom/maelstrom.gen.go 
```

```sh
oapi-codegen -generate types -package maelstrom http/maelstrom.yaml > cmd/maelstrom/internal/maelstrom/types.gen.go 
```