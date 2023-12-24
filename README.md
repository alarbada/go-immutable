# go-immutable

A linter and naming convention to enforce immutability by default in go programs.

# What does this do? 


```go

func test1() {
	a := 1
	a = 2 // this will trigger an error, variables are immutable by default

	_ = a
}


func test2() {
	mutA := 1
	mutA = 2 // this will not, because it is declared as mutable

	_ = mutA
}


// Declaring mutArg as (mut)able, 
func mutateVariable(mutArg *int) {
	*mutArg = 5
}

func test3() {
	variable := 3
	mutVariable := 3
	mutateVariable(&variable) // this will trigger an error
	mutateVariable(&mutVariable) // this will not
}


type User struct {
	Name string
}

func (mutUser *User) ChangeName() {
	mutUser.Name = "John"
}

func test4() {
	user go:= User{}
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

type Manager struct {
	Name string
}

func (mutManager *Manager) ChangeName() {
	mutManager.Name = "John"
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


	manager := Manager{}

	go manager.ChangeName() // this will also trigger an error, user is passed as a mutable reference


	name := "John"

	go func () {
		// this will not trigger an error, because name is immutable
		fmt.Println(name)
	}()
}

```
