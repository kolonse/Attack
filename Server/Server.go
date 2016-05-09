// Server project Server.go
package Server

import (
	"github.com/kolonse/KolonseWeb"
	//	"github.com/kolonse/KolonseWeb/HttpLib"
	//	"github.com/kolonse/KolonseWeb/Type"
	"github.com/kolonse/KolonseBlogWeb/Routes"
	"github.com/kolonse/KolonseWeb/middleWare/StaticDir"
)

func main() {
	KolonseWeb.DefaultApp.Use(StaticDir.NewMiddleWare("Views"))
	Routes.Load(KolonseWeb.DefaultApp)
	KolonseWeb.DefaultApp.Listen("0.0.0.0", *Port)
}
