package httpapi.authz

default allow = false

# Set "input.attributes.request.http" to a variable "request"
request := input.attributes.request.http

# Allow only GET request on /allowed path
allow {
    request.method == "GET"
    request.path == "/allowed"
}
