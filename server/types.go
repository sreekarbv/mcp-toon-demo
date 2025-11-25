package main

type User struct {
	ID   int    `toon:"id" json:"id"`
	Name string `toon:"name" json:"name"`
	Role string `toon:"role" json:"role"`
	City string `toon:"city" json:"city"`
}

type UsersPayload struct {
	Users []User `toon:"users" json:"users"`
}
