/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package api

import (
	"net/http"
	"strconv"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

// UserAPI handles request to /api/users/{}
type UserAPI struct {
	BaseAPI
	currentUserID int
	userID        int
}

// Prepare validates the URL and parms
func (ua *UserAPI) Prepare() {

	ua.currentUserID = ua.ValidateUser()
	id := ua.Ctx.Input.Param(":id")
	if id == "current" {
		ua.userID = ua.currentUserID
	} else if len(id) > 0 {
		var err error
		ua.userID, err = strconv.Atoi(id)
		if err != nil {
			log.Errorf("Invalid user id, error: %v", err)
			ua.CustomAbort(http.StatusBadRequest, "Invalid user Id")
		}
		userQuery := models.User{UserID: ua.userID}
		u, err := dao.GetUser(userQuery)
		if err != nil {
			log.Errorf("Error occurred in GetUser, error: %v", err)
			ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		if u == nil {
			log.Errorf("User with Id: %d does not exist", ua.userID)
			ua.CustomAbort(http.StatusNotFound, "")
		}
	}
}

// Get ...
func (ua *UserAPI) Get() {
	exist, err := dao.IsAdminRole(ua.currentUserID)
	if err != nil {
		log.Errorf("Error occurred in IsAdminRole, error: %v", err)
		ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}

	if ua.userID == 0 { //list users
		if !exist {
			log.Errorf("Current user, id: %d does not have admin role, can not list users", ua.currentUserID)
			ua.RenderError(http.StatusForbidden, "User does not have admin role")
			return
		}
		username := ua.GetString("username")
		userQuery := models.User{}
		if len(username) > 0 {
			userQuery.Username = "%" + username + "%"
		}
		userList, err := dao.ListUsers(userQuery)
		if err != nil {
			log.Errorf("Failed to get data from database, error: %v", err)
			ua.RenderError(http.StatusInternalServerError, "Failed to query from database")
			return
		}
		ua.Data["json"] = userList

	} else if ua.userID == ua.currentUserID || exist {
		userQuery := models.User{UserID: ua.userID}
		u, err := dao.GetUser(userQuery)
		if err != nil {
			log.Errorf("Error occurred in GetUser, error: %v", err)
			ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		ua.Data["json"] = u
	} else {
		log.Errorf("Current user, id: %d does not have admin role, can not view other user's detail", ua.currentUserID)
		ua.RenderError(http.StatusForbidden, "User does not have admin role")
		return
	}
	ua.ServeJSON()
}

// Put ...
func (ua *UserAPI) Put() { //currently only for toggle admin, so no request body
	exist, err := dao.IsAdminRole(ua.currentUserID)
	if err != nil {
		log.Errorf("Error occurred in IsAdminRole, error: %v", err)
		ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if !exist {
		log.Warningf("current user, id: %d does not have admin role, can not update other user's role", ua.currentUserID)
		ua.RenderError(http.StatusForbidden, "User does not have admin role")
		return
	}
	userQuery := models.User{UserID: ua.userID}
	dao.ToggleUserAdminRole(userQuery)
}

// Delete ...
func (ua *UserAPI) Delete() {
	exist, err := dao.IsAdminRole(ua.currentUserID)
	if err != nil {
		log.Errorf("Error occurred in IsAdminRole, error: %v", err)
		ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if !exist {
		log.Warningf("current user, id: %d does not have admin role, can not remove user", ua.currentUserID)
		ua.RenderError(http.StatusForbidden, "User does not have admin role")
		return
	}
	err = dao.DeleteUser(ua.userID)
	if err != nil {
		log.Errorf("Failed to delete data from database, error: %v", err)
		ua.RenderError(http.StatusInternalServerError, "Failed to delete User")
		return
	}
}
