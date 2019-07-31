// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package misc

import (
	"github.com/masoodkamyab/gitea/modules/context"
	"github.com/masoodkamyab/gitea/modules/setting"
	"github.com/masoodkamyab/gitea/modules/structs"
)

// Version shows the version of the Gitea server
func Version(ctx *context.APIContext) {
	// swagger:operation GET /version miscellaneous getVersion
	// ---
	// summary: Returns the version of the Gitea application
	// produces:
	// - application/json
	// responses:
	//   "200":
	//     "$ref": "#/responses/ServerVersion"
	ctx.JSON(200, &structs.ServerVersion{Version: setting.AppVer})
}
