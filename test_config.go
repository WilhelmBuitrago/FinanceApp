package main
import (
	"fmt"
	"financeapp/backend/internal/config"
)
func main(){
	cfg:=config.Load()
	fmt.Println(cfg.DatabasePath)
	fmt.Println(cfg.ArtifactsPath)
	fmt.Println(cfg.LogsPath)
	fmt.Println(cfg.Port)
}
