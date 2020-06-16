package rest

// Controller is a REST action controller interface
type Controller interface {
	// The index action handles index/list requests; it should respond with a
	// list of the requested resources
	Index() error

	//The get action handles GET requests and receives an 'id' parameter; it
	// should respond with the server resource state of the resource identified
	// by the 'id' value
	Get() error

	// The head action handles HEAD requests and receives an 'id' parameter; it
	// should respond with the server resource state of the resource identified
	// by the 'id' value
	Head() error

	// The post action handles POST requests; it should accept and digest a
	// POSTed resource representation and persist the resource state
	Post() error

	// The put action handles PUT requests and receives an 'id' parameter; it
	// should update the server resource state of the resource identified by
	// the 'id' value
	Put() error

	// The delete action handles DELETE requests and receives an 'id'
	// parameter; it should update the server resource state of the resource
	// identified by the 'id' value
	Delete() error
}
