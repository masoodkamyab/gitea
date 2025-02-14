package repo

import (
	"fmt"

	"github.com/masoodkamyab/gitea/models"
	"github.com/masoodkamyab/gitea/modules/context"
	"github.com/masoodkamyab/gitea/modules/git"
)

// SetEditorconfigIfExists set editor config as render variable
func SetEditorconfigIfExists(ctx *context.Context) {
	ec, err := ctx.Repo.GetEditorconfig()

	if err != nil && !git.IsErrNotExist(err) {
		description := fmt.Sprintf("Error while getting .editorconfig file: %v", err)
		if err := models.CreateRepositoryNotice(description); err != nil {
			ctx.ServerError("ErrCreatingReporitoryNotice", err)
		}
		return
	}

	ctx.Data["Editorconfig"] = ec
}

// SetDiffViewStyle set diff style as render variable
func SetDiffViewStyle(ctx *context.Context) {
	queryStyle := ctx.Query("style")

	if !ctx.IsSigned {
		ctx.Data["IsSplitStyle"] = queryStyle == "split"
		return
	}

	var (
		userStyle = ctx.User.DiffViewStyle
		style     string
	)

	if queryStyle == "unified" || queryStyle == "split" {
		style = queryStyle
	} else if userStyle == "unified" || userStyle == "split" {
		style = userStyle
	} else {
		style = "unified"
	}

	ctx.Data["IsSplitStyle"] = style == "split"
	if err := ctx.User.UpdateDiffViewStyle(style); err != nil {
		ctx.ServerError("ErrUpdateDiffViewStyle", err)
	}
}

// SetWhitespaceBehavior set whitespace behavior as render variable
func SetWhitespaceBehavior(ctx *context.Context) {
	whitespaceBehavior := ctx.Query("whitespace")
	switch whitespaceBehavior {
	case "ignore-all", "ignore-eol", "ignore-change":
		ctx.Data["WhitespaceBehavior"] = whitespaceBehavior
	default:
		ctx.Data["WhitespaceBehavior"] = ""
	}
}
