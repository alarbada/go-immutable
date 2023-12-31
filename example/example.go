package example

import (
	"fmt"
	"go-immutable/example/mod1"
)

func (User) ChangeSomething(mutFuncVar int) {
}

func changeUser(u User, mutLolWhat int) {
	mutLolWhat = 4

	u.Name = "John"
}

func test() {
	immutableVariable := 3
	mutableVariable := 4

	changeLol(immutableVariable, mutableVariable)

	mod1.ChangeSomething(immutableVariable)

	var lol mod1.LolWhat
	lol.ChangeSomething(immutableVariable)

	mutVariable := immutableVariable
	var user User
	user.ChangeSomething(mutVariable)
}

func mutateVariable(arg *int) {
	*arg = 5
}

func test3() {
	variable := 3
	mutateVariable(&variable)
}

type User struct {
	Name string
}

func (mutUser *User) ChangeName() {
	mutUser.Name = "John"
}

func test4() {
	user := User{}
	mutUser := User{}

	user.ChangeName()    // This triggers an error
	mutUser.ChangeName() // This will not
}

func test5() {
	// Same for slices and maps

	mutSlice := []int{1, 2, 3}
	mutSlice[0] = 4 // This will not trigger an error

	immutableMap := map[string]int{"a": 1, "b": 2}
	immutableMap["a"] = 3 // This will trigger an error
}


func test6() {
	// goodbye to unintended mutations

	doSomething := func() (int, int, error) {
		return 1, 2, nil
	}

	a := 3

	result, a, err := doSomething() // This will trigger an error, a is immutable

	_ = a
	_ = result
	_ = err

	// err is a special case, it is always mutable so that we don't break the usual golang error handling
	_, _, err = doSomething() // This will not trigger an error
}

func test7() {
	mutName := "John"

	go func () {
		// at the moment this will always trigger an error. You can't share mutable variables between goroutines, you should use immutable values (without the mut prefix)
		mutName = "Jane"
	}()
	_ = mutName

	go func (mutName *string) {
		// Not even this will work, mutName is still mutable
		*mutName = "Jane"
	}(&mutName)
	_ = mutName


	user := User{}

	go user.ChangeName() // this will also trigger an error, user is passed as a mutable reference


	name := "John"

	go func () {
		// this will not trigger an error, because name is immutable
		fmt.Println(name)
	}()
}
