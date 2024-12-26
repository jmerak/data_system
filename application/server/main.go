package main

import (
	"server/api"
	"server/routers"
)

func main() {
	r := routers.SetupRouter()
	api.Init()
	r.Run(":9090")
}
