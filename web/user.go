package web

import (
	"context"
	"database/sql"
	"errors"
	"go-rest-api/api/model"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type CreateUserRequest struct {
	// allow empty
	UserName string `json:"username"`
}

func (req *CreateUserRequest) Bind(r *http.Request) error {
	err := validate.Struct(*req)
	return err
}

type UpdateUserRequest struct {
	// allow empty
	UserName *string `json:"username"`
}

func (req *UpdateUserRequest) Bind(r *http.Request) error {
	err := validate.Struct(*req)
	return err
}

type UserResponse struct {
	ID       string `json:"id"`
	UserName string `json:"username"`
}

func NewUserResponse(user model.User) UserResponse {
	result := UserResponse{
		ID:       user.PublicID,
		UserName: user.Username,
	}
	return result
}

func NewUsersResponse(users []model.User) []UserResponse {
	result := []UserResponse{}
	for _, u := range users {
		result = append(result, NewUserResponse(u))
	}
	return result
}

// GET /users/
func (api *API) GetAllUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, err := api.Repo.GetUsers(ctx)
	if err != nil {
		panic(err)
	}
	RenderJSON(w, http.StatusOK, NewUsersResponse(users))
}

// POST /users/
func (api *API) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var data CreateUserRequest
	if err := render.Bind(r, &data); err != nil {
		panic(ErrValidation(err))
	}

	id, err := api.Repo.CreateUser(ctx, data.UserName)
	if err != nil {
		panic(err)
	}

	u, err := api.Repo.GetUserByID(ctx, id)
	if err != nil {
		panic(err)
	}
	RenderJSON(w, http.StatusCreated, NewUserResponse(u))
}

// GET /users/{id}
func (api *API) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u, _ := pathUserFrom(ctx)
	RenderJSON(w, http.StatusOK, NewUserResponse(u))
}

// PATCH /users/{id}
func (api *API) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u, _ := pathUserFrom(ctx)
	var data UpdateUserRequest
	if err := render.Bind(r, &data); err != nil {
		panic(ErrValidation(err))
	}

	if data.UserName != nil {
		u.Username = *data.UserName
	}

	_, err := api.Repo.UpdateUserByID(ctx, u)
	if err != nil {
		panic(err)
	}

	u, err = api.Repo.GetUserByPublicID(ctx, u.PublicID)
	if err != nil {
		panic(err)
	}
	RenderJSON(w, http.StatusOK, NewUserResponse(u))
}

// DELETE /users/{id}
func (api *API) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u, _ := pathUserFrom(ctx)

	_, err := api.Repo.DeleteUserByID(ctx, u.ID)
	if err != nil {
		panic(err)
	}
	RenderNoContent(w)
}

func (api *API) RequireUserID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		pathUserID := chi.URLParam(r, "id")
		pathUser, err := api.Repo.GetUserByPublicID(ctx, pathUserID)
		if errors.Is(err, sql.ErrNoRows) {
			panic(ErrValidationWithMessage(err, ErrMessageNotFoundUser))
		} else if err != nil {
			panic(err)
		}
		ctx = withPathUser(ctx, pathUser)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

type pathUserCtxKey struct{}

func withPathUser(ctx context.Context, pathUser model.User) context.Context {
	return context.WithValue(ctx, pathUserCtxKey{}, pathUser)
}

func pathUserFrom(ctx context.Context) (model.User, bool) {
	pathUser, ok := ctx.Value(pathUserCtxKey{}).(model.User)
	return pathUser, ok
}
