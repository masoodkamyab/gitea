// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package org

import (
	"net/http"
	"path"
	"strings"

	"github.com/Unknwon/com"

	"github.com/masoodkamyab/gitea/models"
	"github.com/masoodkamyab/gitea/modules/auth"
	"github.com/masoodkamyab/gitea/modules/base"
	"github.com/masoodkamyab/gitea/modules/context"
	"github.com/masoodkamyab/gitea/modules/log"
	"github.com/masoodkamyab/gitea/routers/utils"
)

const (
	// tplTeams template path for teams list page
	tplTeams base.TplName = "org/team/teams"
	// tplTeamNew template path for create new team page
	tplTeamNew base.TplName = "org/team/new"
	// tplTeamMembers template path for showing team members page
	tplTeamMembers base.TplName = "org/team/members"
	// tplTeamRepositories template path for showing team repositories page
	tplTeamRepositories base.TplName = "org/team/repositories"
)

// Teams render teams list page
func Teams(ctx *context.Context) {
	org := ctx.Org.Organization
	ctx.Data["Title"] = org.FullName
	ctx.Data["PageIsOrgTeams"] = true

	for _, t := range org.Teams {
		if err := t.GetMembers(); err != nil {
			ctx.ServerError("GetMembers", err)
			return
		}
	}
	ctx.Data["Teams"] = org.Teams

	ctx.HTML(200, tplTeams)
}

// TeamsAction response for join, leave, remove, add operations to team
func TeamsAction(ctx *context.Context) {
	uid := com.StrTo(ctx.Query("uid")).MustInt64()
	if uid == 0 {
		ctx.Redirect(ctx.Org.OrgLink + "/teams")
		return
	}

	page := ctx.Query("page")
	var err error
	switch ctx.Params(":action") {
	case "join":
		if !ctx.Org.IsOwner {
			ctx.Error(404)
			return
		}
		err = ctx.Org.Team.AddMember(ctx.User.ID)
	case "leave":
		err = ctx.Org.Team.RemoveMember(ctx.User.ID)
	case "remove":
		if !ctx.Org.IsOwner {
			ctx.Error(404)
			return
		}
		err = ctx.Org.Team.RemoveMember(uid)
		page = "team"
	case "add":
		if !ctx.Org.IsOwner {
			ctx.Error(404)
			return
		}
		uname := utils.RemoveUsernameParameterSuffix(strings.ToLower(ctx.Query("uname")))
		var u *models.User
		u, err = models.GetUserByName(uname)
		if err != nil {
			if models.IsErrUserNotExist(err) {
				ctx.Flash.Error(ctx.Tr("form.user_not_exist"))
				ctx.Redirect(ctx.Org.OrgLink + "/teams/" + ctx.Org.Team.LowerName)
			} else {
				ctx.ServerError(" GetUserByName", err)
			}
			return
		}

		if u.IsOrganization() {
			ctx.Flash.Error(ctx.Tr("form.cannot_add_org_to_team"))
			ctx.Redirect(ctx.Org.OrgLink + "/teams/" + ctx.Org.Team.LowerName)
			return
		}

		if ctx.Org.Team.IsMember(u.ID) {
			ctx.Flash.Error(ctx.Tr("org.teams.add_duplicate_users"))
		} else {
			err = ctx.Org.Team.AddMember(u.ID)
		}

		page = "team"
	}

	if err != nil {
		if models.IsErrLastOrgOwner(err) {
			ctx.Flash.Error(ctx.Tr("form.last_org_owner"))
		} else {
			log.Error("Action(%s): %v", ctx.Params(":action"), err)
			ctx.JSON(200, map[string]interface{}{
				"ok":  false,
				"err": err.Error(),
			})
			return
		}
	}

	switch page {
	case "team":
		ctx.Redirect(ctx.Org.OrgLink + "/teams/" + ctx.Org.Team.LowerName)
	case "home":
		ctx.Redirect(ctx.Org.Organization.HomeLink())
	default:
		ctx.Redirect(ctx.Org.OrgLink + "/teams")
	}
}

// TeamsRepoAction operate team's repository
func TeamsRepoAction(ctx *context.Context) {
	if !ctx.Org.IsOwner {
		ctx.Error(404)
		return
	}

	var err error
	switch ctx.Params(":action") {
	case "add":
		repoName := path.Base(ctx.Query("repo_name"))
		var repo *models.Repository
		repo, err = models.GetRepositoryByName(ctx.Org.Organization.ID, repoName)
		if err != nil {
			if models.IsErrRepoNotExist(err) {
				ctx.Flash.Error(ctx.Tr("org.teams.add_nonexistent_repo"))
				ctx.Redirect(ctx.Org.OrgLink + "/teams/" + ctx.Org.Team.LowerName + "/repositories")
				return
			}
			ctx.ServerError("GetRepositoryByName", err)
			return
		}
		err = ctx.Org.Team.AddRepository(repo)
	case "remove":
		err = ctx.Org.Team.RemoveRepository(com.StrTo(ctx.Query("repoid")).MustInt64())
	}

	if err != nil {
		log.Error("Action(%s): '%s' %v", ctx.Params(":action"), ctx.Org.Team.Name, err)
		ctx.ServerError("TeamsRepoAction", err)
		return
	}
	ctx.Redirect(ctx.Org.OrgLink + "/teams/" + ctx.Org.Team.LowerName + "/repositories")
}

// NewTeam render create new team page
func NewTeam(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Org.Organization.FullName
	ctx.Data["PageIsOrgTeams"] = true
	ctx.Data["PageIsOrgTeamsNew"] = true
	ctx.Data["Team"] = &models.Team{}
	ctx.Data["Units"] = models.Units
	ctx.HTML(200, tplTeamNew)
}

// NewTeamPost response for create new team
func NewTeamPost(ctx *context.Context, form auth.CreateTeamForm) {
	ctx.Data["Title"] = ctx.Org.Organization.FullName
	ctx.Data["PageIsOrgTeams"] = true
	ctx.Data["PageIsOrgTeamsNew"] = true
	ctx.Data["Units"] = models.Units

	t := &models.Team{
		OrgID:       ctx.Org.Organization.ID,
		Name:        form.TeamName,
		Description: form.Description,
		Authorize:   models.ParseAccessMode(form.Permission),
	}

	if t.Authorize < models.AccessModeOwner {
		var units = make([]*models.TeamUnit, 0, len(form.Units))
		for _, tp := range form.Units {
			units = append(units, &models.TeamUnit{
				OrgID: ctx.Org.Organization.ID,
				Type:  tp,
			})
		}
		t.Units = units
	}

	ctx.Data["Team"] = t

	if ctx.HasError() {
		ctx.HTML(200, tplTeamNew)
		return
	}

	if t.Authorize < models.AccessModeAdmin && len(form.Units) == 0 {
		ctx.RenderWithErr(ctx.Tr("form.team_no_units_error"), tplTeamNew, &form)
		return
	}

	if err := models.NewTeam(t); err != nil {
		ctx.Data["Err_TeamName"] = true
		switch {
		case models.IsErrTeamAlreadyExist(err):
			ctx.RenderWithErr(ctx.Tr("form.team_name_been_taken"), tplTeamNew, &form)
		default:
			ctx.ServerError("NewTeam", err)
		}
		return
	}
	log.Trace("Team created: %s/%s", ctx.Org.Organization.Name, t.Name)
	ctx.Redirect(ctx.Org.OrgLink + "/teams/" + t.LowerName)
}

// TeamMembers render team members page
func TeamMembers(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Org.Team.Name
	ctx.Data["PageIsOrgTeams"] = true
	ctx.Data["PageIsOrgTeamMembers"] = true
	if err := ctx.Org.Team.GetMembers(); err != nil {
		ctx.ServerError("GetMembers", err)
		return
	}
	ctx.HTML(200, tplTeamMembers)
}

// TeamRepositories show the repositories of team
func TeamRepositories(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Org.Team.Name
	ctx.Data["PageIsOrgTeams"] = true
	ctx.Data["PageIsOrgTeamRepos"] = true
	if err := ctx.Org.Team.GetRepositories(); err != nil {
		ctx.ServerError("GetRepositories", err)
		return
	}
	ctx.HTML(200, tplTeamRepositories)
}

// EditTeam render team edit page
func EditTeam(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Org.Organization.FullName
	ctx.Data["PageIsOrgTeams"] = true
	ctx.Data["team_name"] = ctx.Org.Team.Name
	ctx.Data["desc"] = ctx.Org.Team.Description
	ctx.Data["Units"] = models.Units
	ctx.HTML(200, tplTeamNew)
}

// EditTeamPost response for modify team information
func EditTeamPost(ctx *context.Context, form auth.CreateTeamForm) {
	t := ctx.Org.Team
	ctx.Data["Title"] = ctx.Org.Organization.FullName
	ctx.Data["PageIsOrgTeams"] = true
	ctx.Data["Team"] = t
	ctx.Data["Units"] = models.Units

	isAuthChanged := false
	if !t.IsOwnerTeam() {
		// Validate permission level.
		auth := models.ParseAccessMode(form.Permission)

		t.Name = form.TeamName
		if t.Authorize != auth {
			isAuthChanged = true
			t.Authorize = auth
		}
	}
	t.Description = form.Description
	if t.Authorize < models.AccessModeOwner {
		var units = make([]models.TeamUnit, 0, len(form.Units))
		for _, tp := range form.Units {
			units = append(units, models.TeamUnit{
				OrgID:  t.OrgID,
				TeamID: t.ID,
				Type:   tp,
			})
		}
		err := models.UpdateTeamUnits(t, units)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, "LoadIssue", err.Error())
			return
		}
	}

	if ctx.HasError() {
		ctx.HTML(200, tplTeamNew)
		return
	}

	if t.Authorize < models.AccessModeAdmin && len(form.Units) == 0 {
		ctx.RenderWithErr(ctx.Tr("form.team_no_units_error"), tplTeamNew, &form)
		return
	}

	if err := models.UpdateTeam(t, isAuthChanged); err != nil {
		ctx.Data["Err_TeamName"] = true
		switch {
		case models.IsErrTeamAlreadyExist(err):
			ctx.RenderWithErr(ctx.Tr("form.team_name_been_taken"), tplTeamNew, &form)
		default:
			ctx.ServerError("UpdateTeam", err)
		}
		return
	}
	ctx.Redirect(ctx.Org.OrgLink + "/teams/" + t.LowerName)
}

// DeleteTeam response for the delete team request
func DeleteTeam(ctx *context.Context) {
	if err := models.DeleteTeam(ctx.Org.Team); err != nil {
		ctx.Flash.Error("DeleteTeam: " + err.Error())
	} else {
		ctx.Flash.Success(ctx.Tr("org.teams.delete_team_success"))
	}

	ctx.JSON(200, map[string]interface{}{
		"redirect": ctx.Org.OrgLink + "/teams",
	})
}
