package routers

import (
	"github.com/masoodkamyab/gitea/modules/base"
	"github.com/masoodkamyab/gitea/modules/context"
)

// tplSwaggerV1Json swagger v1 json template
const tplSwaggerV1Json base.TplName = "swagger/v1_json"

// SwaggerV1Json render swagger v1 json
func SwaggerV1Json(ctx *context.Context) {
	ctx.HTML(200, tplSwaggerV1Json)
}
