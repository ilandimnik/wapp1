package main

import (
  "github.com/kylelemons/go-gypsy/yaml"
  "log"
  "strings"
)

var config = map[string]string{}

//Currently on init
func init() {
  config["auth_key"] = "620b9b69125d5b3dd93b38c9038386d712284528c148aa23af40094631680e5d6c45ef8fdbe272a28ee38caffc237f3baabce4f81ca8c4bbc46950ebf3a92e9a"
  config["enc_key"] = "da445212bcbc76abde79bbbf7beff3cf235d7b6005fa721a7a438d5a914c2c40"

  cfg_file := "config/database.yaml"
  yaml_config, err := yaml.ReadFile(cfg_file)
  if err != nil {
    log.Fatalf("readfile(%q): %s", cfg_file, err)
  }

  params := []string{"db_url", "db_name"}

  for _, param := range params {
    val, err := yaml_config.Get("development." + param)
    if err != nil {
      log.Printf("params: %s, err: %s\n", param, err)
      continue
    }
    config[param] = strings.TrimSpace(val)
  }
}
