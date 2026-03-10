package api

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/kgretzky/pwndrop/storage"
)

func UsersOptionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
}

func UserListHandler(w http.ResponseWriter, r *http.Request) {
	// #### CHECK IF AUTHENTICATED ####
	_, err := AuthSession(r)
	if err != nil {
		DumpResponse(w, "unauthorized", http.StatusUnauthorized, API_ERROR_BAD_AUTHENTICATION, nil)
		return
	}

	type UserItem struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	type UserListResponse struct {
		Users []UserItem `json:"users"`
	}

	users, err := storage.UserList()
	if err != nil {
		DumpResponse(w, err.Error(), http.StatusInternalServerError, API_ERROR_FILE_DATABASE_FAILED, nil)
		return
	}

	resp := &UserListResponse{}
	for _, user := range users {
		resp.Users = append(resp.Users, UserItem{
			ID:   user.ID,
			Name: user.Name,
		})
	}

	DumpResponse(w, "ok", http.StatusOK, 0, resp)
}

func UserDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// #### CHECK IF AUTHENTICATED ####
	current_uid, err := AuthSession(r)
	if err != nil {
		DumpResponse(w, "unauthorized", http.StatusUnauthorized, API_ERROR_BAD_AUTHENTICATION, nil)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		DumpResponse(w, err.Error(), http.StatusBadRequest, API_ERROR_BAD_REQUEST, nil)
		return
	}

	users, err := storage.UserList()
	if err != nil {
		DumpResponse(w, err.Error(), http.StatusInternalServerError, API_ERROR_FILE_DATABASE_FAILED, nil)
		return
	}
	if len(users) <= 1 {
		DumpResponse(w, "can't delete the last user", http.StatusBadRequest, API_ERROR_BAD_REQUEST, nil)
		return
	}

	_, err = storage.UserGet(id)
	if err != nil {
		DumpResponse(w, err.Error(), http.StatusNotFound, API_ERROR_BAD_REQUEST, nil)
		return
	}

	err = storage.UserDelete(id)
	if err != nil {
		DumpResponse(w, err.Error(), http.StatusInternalServerError, API_ERROR_FILE_DATABASE_FAILED, nil)
		return
	}

	err = storage.SessionDeleteByUid(id)
	if err != nil {
		DumpResponse(w, err.Error(), http.StatusInternalServerError, API_ERROR_FILE_DATABASE_FAILED, nil)
		return
	}

	if current_uid == id {
		deleteCookie(AUTH_COOKIE_NAME, w)
	}

	DumpResponse(w, "ok", http.StatusOK, 0, nil)
}
