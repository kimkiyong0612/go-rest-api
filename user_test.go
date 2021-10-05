package main

import (
	"net/http"

	"github.com/gavv/httpexpect/v2"
)

//  referenced doc:https://github.com/gavv/httpexpect/blob/master/_examples/echo_test.go
// TODO: add more testcase
func UserTest(e *httpexpect.Expect) {
	//create user
	resp := e.POST("/v1/users").
		WithJSON(map[string]string{
			"username": "sample_user",
		}).Expect().Status(http.StatusCreated).JSON().Object()

	// check
	resp.ValueEqual("username", "sample_user")
	resp.Keys().ContainsOnly("id", "username")

	user1 := resp.Raw()

	// get user
	e.GET("/v1/users/" + user1["id"].(string)).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	// check
	resp.ValueEqual("username", "sample_user")
	resp.Keys().ContainsOnly("id", "username")

	//get users
	e.GET("/v1/users").
		WithJSON(map[string]string{
			"username": "sample_user",
		}).Expect().Status(http.StatusOK)

	// update user
	resp = e.PATCH("/v1/users/" + user1["id"].(string)).
		WithJSON(map[string]string{
			"username": "updated_user",
		}).
		Expect().Status(http.StatusOK).JSON().Object()

	resp.ValueEqual("username", "updated_user")
	resp.Keys().ContainsOnly("id", "username")

	// delete user
	e.DELETE("/v1/users/" + user1["id"].(string)).
		Expect().Status(http.StatusNoContent)
}
