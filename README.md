# Maelstrom

Playing around with hex architecture, htmx and gossip for a distributed load test tool. Is this project complete? Absolutely not. Do I remember where this code base when I last used it, no. 
Should you use it as is, no. But I had fun building it


```sh
oapi-codegen -generate chi-server -package maelstrom http/maelstrom.yaml > cmd/maelstrom/internal/maelstrom/maelstrom.gen.go 
```

```sh
oapi-codegen -generate types -package maelstrom http/maelstrom.yaml > cmd/maelstrom/internal/maelstrom/types.gen.go 
```
