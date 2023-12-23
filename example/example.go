package example

import "go-immutable/example/mod1"

type User struct {
	mutName string
	mutAge  int
}

func (User) ChangeSomething(mutFuncVar int) {
}

func changeUser(u User, mutLolWhat int) {
	mutLolWhat = 4

	u.mutName = "John"
	u.mutAge = 30
}

func test() {
	immutableVariable := 3
	mutableVariable := 4

	changeLol(immutableVariable, mutableVariable)

	mod1.ChangeSomething(immutableVariable)

	var lol mod1.LolWhat
	lol.ChangeSomething(immutableVariable)


	var user User
	user.ChangeSomething(immutableVariable)
}
