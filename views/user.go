package views

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ystv/web-auth/public/templates"
	"github.com/ystv/web-auth/user"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type (
	// UsersTemplate represents the context for the user template
	UsersTemplate struct {
		Users                                 []UserStripped
		UserID                                int
		CurPage, NextPage, PrevPage, LastPage int
		ActivePage                            string
		Sort                                  struct {
			Pages      int
			Size       int
			PageNumber int
			Column     string
			Direction  string
			Search     string
		}
	}
)

// UsersFunc handles a users request
func (v *Views) UsersFunc(w http.ResponseWriter, r *http.Request) {
	session, _ := v.cookie.Get(r, v.conf.SessionCookieName)

	c := v.getData(session)

	var err error

	if r.Method == "POST" {
		err = r.ParseForm()
		if err != nil {
			err = v.errorHandle(w, err)
			if err != nil {
				log.Println(err)
				return
			}
		}

		column := r.FormValue("column")
		direction := r.FormValue("direction")
		search := r.FormValue("search")
		var size int
		sizeRaw := r.FormValue("size")
		if sizeRaw == "all" {
			size = 0
		} else {
			size, err = strconv.Atoi(sizeRaw)
			if err != nil {
				size = 0
			} else if size <= 0 {
				err = v.errorHandle(w, fmt.Errorf("invalid size, must be positive"))
				if err != nil {
					log.Println(err)
					return
				}
			} else if size != 5 && size != 10 && size != 25 && size != 50 && size != 75 && size != 100 {
				size = 25
			}
		}
		valid := true
		switch column {
		case "userId":
		case "name":
		case "username":
		case "email":
		case "lastLogin":
			break
		default:
			valid = false
		}
		switch direction {
		case "asc":
		case "desc":
			break
		default:
			valid = false
		}
		r.Method = "GET"
		if valid {
			if len(search) > 0 {
				if size > 0 {
					http.Redirect(w, r, fmt.Sprintf("/internal/users?column=%s&direction=%s&search=%s&size=%d&page=1", column, direction, url.QueryEscape(search), size), http.StatusFound)
				} else {
					http.Redirect(w, r, fmt.Sprintf("/internal/users?column=%s&direction=%s&search=%s", column, direction, url.QueryEscape(search)), http.StatusFound)
				}
			} else {
				if size > 0 {
					http.Redirect(w, r, fmt.Sprintf("/internal/users?column=%s&direction=%s&size=%d&page=1", column, direction, size), http.StatusFound)
				} else {
					http.Redirect(w, r, fmt.Sprintf("/internal/users?column=%s&direction=%s", column, direction), http.StatusFound)
				}
			}
		} else if len(search) > 0 {
			if size > 0 {
				http.Redirect(w, r, fmt.Sprintf("/internal/users?search=%s&size=%d&page=1", url.QueryEscape(search), size), http.StatusFound)
			} else {
				http.Redirect(w, r, fmt.Sprintf("/internal/users?search=%s", url.QueryEscape(search)), http.StatusFound)
			}
		} else {
			if size > 0 {
				http.Redirect(w, r, fmt.Sprintf("/internal/users?size=%d&page=1", size), http.StatusFound)
			} else {
				r.URL.Query().Del("*")
				http.Redirect(w, r, "/internal/users", http.StatusFound)
			}
		}
		return
	}

	column := r.URL.Query().Get("column")
	direction := r.URL.Query().Get("direction")
	search := r.URL.Query().Get("search")
	var size, page, count int
	sizeRaw := r.URL.Query().Get("size")
	if sizeRaw == "all" {
		size = 0
	} else if len(sizeRaw) != 0 {
		page, err = strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			page = 1
			log.Println(err)
			err = v.errorHandle(w, err)
			if err != nil {
				log.Println(err)
				return
			}
		}
		size, err = strconv.Atoi(sizeRaw)
		if err != nil {
			size = 0
		} else if size <= 0 {
			err = v.errorHandle(w, fmt.Errorf("invalid size, must be positive"))
			if err != nil {
				log.Println(err)
				return
			}
		} else if size != 5 && size != 10 && size != 25 && size != 50 && size != 75 && size != 100 {
			size = 0
		}

		count, err = v.user.CountUsers(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if count <= size*(page-1) {
			log.Println("size and page given is not valid")
			http.Error(w, fmt.Sprintln("size and page given is not valid"), http.StatusBadRequest)
			return
		}
	}

	valid := true
	switch column {
	case "userId":
	case "name":
	case "username":
	case "email":
	case "lastLogin":
		break
	default:
		valid = false
	}
	switch direction {
	case "asc":
	case "desc":
		break
	default:
		valid = false
	}
	var dbUsers []user.User
	if valid {
		if len(search) > 0 {
			if size > 0 && page > 0 {
				dbUsers, err = v.user.GetUsersSortedSearchSizePage(r.Context(), column, direction, search, size, page)
				if err != nil {
					log.Println(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				tmp, err := v.user.GetUsersSortedSearch(r.Context(), column, direction, search)
				if err != nil {
					log.Println(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				count = len(tmp)
			} else {
				dbUsers, err = v.user.GetUsersSortedSearch(r.Context(), column, direction, search)
			}
		} else {
			if size > 0 && page > 0 {
				dbUsers, err = v.user.GetUsersSortedSizePage(r.Context(), column, direction, size, page)
			} else {
				dbUsers, err = v.user.GetUsersSorted(r.Context(), column, direction)
			}
		}
	} else if len(search) > 0 {
		if size > 0 && page > 0 {
			dbUsers, err = v.user.GetUsersSearchSizePage(r.Context(), search, size, page)
			tmp, err := v.user.GetUsersSearch(r.Context(), search)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			count = len(tmp)
		} else {
			dbUsers, err = v.user.GetUsersSearch(r.Context(), search)
		}
	} else {
		if size > 0 && page > 0 {
			dbUsers, err = v.user.GetUsersSizePage(r.Context(), size, page)
		} else {
			dbUsers, err = v.user.GetUsers(r.Context())
		}
	}
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tplUsers := DBToTemplateType(&dbUsers)

	var sum int

	if size == 0 {
		sum = 0
	} else {
		sum = int(math.Ceil(float64(count) / float64(size)))
	}

	ctx := UsersTemplate{
		Users:      tplUsers,
		UserID:     c.User.UserID,
		ActivePage: "users",
		Sort: struct {
			Pages      int
			Size       int
			PageNumber int
			Column     string
			Direction  string
			Search     string
		}{
			Pages:      sum,
			Size:       size,
			PageNumber: page,
			Column:     column,
			Direction:  direction,
			Search:     search,
		},
	}
	err = v.template.RenderTemplatePagination(w, ctx, templates.UsersTemplate)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UserFunc handles a users request
func (v *Views) UserFunc(w http.ResponseWriter, r *http.Request) {
	session, _ := v.cookie.Get(r, v.conf.SessionCookieName)

	c := v.getData(session)

	userString := mux.Vars(r)
	userID, err := strconv.Atoi(userString["userid"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user1, err := v.user.GetUser(r.Context(), user.User{UserID: userID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	perms, err := v.user.GetPermissionsForUser(r.Context(), user1)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, perm := range perms {
		user1.Permissions = append(user1.Permissions, perm)
	}

	roles, err := v.user.GetRolesForUser(r.Context(), user1)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, role1 := range roles {
		user1.Roles = append(user1.Roles, role1.Name)
	}

	user1.Permissions = v.removeDuplicate(user1.Permissions)

	var gravatar string

	if user1.UseGravatar {
		hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace("liam.burnand@bswdi.co.uk"))))
		gravatar = fmt.Sprintf("https://www.gravatar.com/avatar/%s", hex.EncodeToString(hash[:]))
	}

	var createdBy, updatedBy, deletedBy user.User

	if user1.CreatedBy.Valid {
		createdBy, err = v.user.GetUser(r.Context(), user.User{UserID: int(user1.CreatedBy.Int64)})
		if err != nil {
			log.Println(err)
			createdBy.UserID = int(user1.CreatedBy.Int64)
			createdBy.Firstname = ""
			createdBy.Lastname = ""
		}
	} else {
		createdBy.UserID = -1
		createdBy.Firstname = ""
		createdBy.Lastname = ""
	}

	if user1.UpdatedBy.Valid {
		updatedBy, err = v.user.GetUser(r.Context(), user.User{UserID: int(user1.UpdatedBy.Int64)})
		if err != nil {
			log.Println(err)
			updatedBy.UserID = int(user1.UpdatedBy.Int64)
			updatedBy.Firstname = ""
			updatedBy.Lastname = ""
		}
	} else {
		updatedBy.UserID = -1
		updatedBy.Firstname = ""
		updatedBy.Lastname = ""
	}

	if user1.DeletedBy.Valid {
		deletedBy, err = v.user.GetUser(r.Context(), user.User{UserID: int(user1.DeletedBy.Int64)})
		if err != nil {
			log.Println(err)
			deletedBy.UserID = int(user1.DeletedBy.Int64)
			deletedBy.Firstname = ""
			deletedBy.Lastname = ""
		}
	} else {
		deletedBy.UserID = -1
		deletedBy.Firstname = ""
		deletedBy.Lastname = ""
	}

	data := struct {
		User       user.User
		UserID     int
		ActivePage string
		LastLogin  string
		Gravatar   string
		CreatedBy  user.User
		CreatedAt  string
		UpdatedBy  user.User
		UpdatedAt  string
		DeletedBy  user.User
		DeletedAt  string
	}{
		ActivePage: "user",
		User:       user1,
		UserID:     c.User.UserID,
		LastLogin:  user1.LastLogin.Time.Format("2006-01-02 15:04:05"),
		Gravatar:   gravatar,
		CreatedBy:  createdBy,
		CreatedAt:  user1.CreatedAt.Time.Format("2006-01-02 15:04:05"),
		UpdatedBy:  updatedBy,
		UpdatedAt:  user1.UpdatedAt.Time.Format("2006-01-02 15:04:05"),
		DeletedBy:  deletedBy,
		DeletedAt:  user1.DeletedAt.Time.Format("2006-01-02 15:04:05"),
	}

	err = v.template.RenderTemplate(w, data, templates.UserTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
