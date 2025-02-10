package example

// Struct is a struct that will be used to generate code.
//
//genex:json
type Struct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

//genex:json
type Second struct {
	Kekus string `json:"kekus"`
}
