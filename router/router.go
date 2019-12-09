/*
 * Nanocloud Community, a comprehensive platform to turn any application
 * into a cloud solution.
 *
 * Copyright (C) 2016 Nanocloud Software
 *
 * This file is part of Nanocloud community.
 *
 * Nanocloud community is free software; you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * Nanocloud community is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package router

import (
	"net/http"
	"os"

	"main/routes/about"
	"main/routes/shells"

	"main/routes/sessions"

	"main/routes/apps"

	"main/routes/exec"

	"main/routes/files"

	"main/routes/power"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	log "github.com/sirupsen/logrus"
)

type hash map[string]interface{}

func Start() {
	e := echo.New()

	e.Use(mw.Recover())

	e.POST("/exec", exec.Route)
	e.GET("/", about.Get)

	/***
	FILES
	***/

	e.GET("/files", files.Get)
	e.PATCH("/files", files.Patch)
	e.DELETE("/files", files.Delete)
	e.POST("/upload", files.Post)
	e.POST("/directory", files.CreateDirectory)

	/***
	POWER
	***/

	e.GET("/shutdown", power.ShutDown)
	e.GET("/restart", power.Restart)
	e.GET("/checkrds", power.CheckRDS)

	/***
	SESSIONS
	***/

	e.GET("/sessions/:id", sessions.Get)
	e.DELETE("/sessions/:id", sessions.Logoff)

	/***
	SHELLS
	***/

	e.POST("/shells", shells.Post)

	/***
	APPS
	***/

	e.POST("/publishapp", apps.PublishApp)
	e.GET("/apps", apps.GetApps)
	e.DELETE("/apps/:id", apps.UnpublishApp)

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		c.JSON(
			http.StatusInternalServerError,
			hash{
				"error": err.Error(),
			},
		)
	}

	port := os.Getenv("PLAZA_PORT")
	if port == "" {
		port = "9090"
	}
	log.Info("Listenning on port: " + port)
	e.Start(":" + port)
}
