package patrol

func (self *Patrol) Ping(
	request *API_Request,
) *API_Response {
	// Ping doesn't support Services
	// pinging service doesn't make sense currently
	return self.api(true, request)
}
