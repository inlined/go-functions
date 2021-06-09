package example

import (
	"context"
	"fmt"
	"net/http"

	"firebase.com/functions"
)

func myFunction(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Called HTTP trigger")
	w.WriteHeader(http.StatusOK)
}

type User struct {
	UID         string `json:"uid"`
	DisplayName string `json:"displayName"`
}

func onSignUp(ctx context.Context, u *User) {
	fmt.Println("onSignUp():", u.UID)
}

var MyFunction = functions.OnRequest(myFunction)

var MyOtherFunction = functions.OnPubSubEvent("test_topic", onSignUp)
