package main

import (
	"net/url"
	"strings"
	"fmt"
	"reflect"
	"net/http"
	"encoding/json"
)

/*
func HelloResponse() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		fmt.Fprintf(w, "hello world")
	}
}

func main() {
	http.HandleFunc("/", HelloResponse())
	http.ListenAndServe(":5599", nil)
}

type IndexController struct {
}

func (IndexController *IndexController) Index() () {

}
*/

/**
	[
		"/"   :["GET":["IndexController":"Get"], "POST":["IndexController":"Post"]],
		"/foo":["GET":["IndexController":"Foo"]]
	]
 */
type APISerivce struct {
	controllerRegistry          map[string]interface{}
	registeredPathAndController map[string]map[string]map[string]string
	requestForm                 map[string]url.Values
}

func (api *APISerivce) Get(path, controllerWithActionString string) {

	mapping := api.mappingRequestMethodWithControllerAndActions("GET", path, controllerWithActionString)

	fmt.Println(mapping)

	api.registeredPathAndController[path] = mapping
}

func (api *APISerivce) Post(path, controllerWithActionString string) {

	mapping := api.mappingRequestMethodWithControllerAndActions("POST", path, controllerWithActionString)

	fmt.Println(mapping)

	api.registeredPathAndController[path] = mapping
}

func (api *APISerivce) mappingRequestMethodWithControllerAndActions(requestMethod, path, controllerWithActionString string) map[string]map[string]string {

	mappingRequest := make(map[string]map[string]string)

	if length := len(api.registeredPathAndController[path]); length > 0 {
		mappingRequest = api.registeredPathAndController[path]
	}

	controllerAndActionSlice := strings.Split(controllerWithActionString, "@");
	controller := controllerAndActionSlice[0]
	action := controllerAndActionSlice[1]

	controllerAndActionMap := map[string]string{controller: action }

	mappingRequest[requestMethod] = controllerAndActionMap

	return mappingRequest
}

func (api *APISerivce) registerController(controller interface{}) {

	var controllerType string

	if t := reflect.TypeOf(controller); t.Kind() == reflect.Ptr {
		controllerType = t.Elem().Name()
	} else {
		controllerType = t.Name();
	}

	api.controllerRegistry[controllerType] = controller

}

func (api *APISerivce) HandleRequest(controller map[string]map[string]string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		method := r.Method
		matchedController, ok := controller[method]

		if !ok {
			w.WriteHeader(405)
		}

		r.ParseForm()
		api.requestForm["query"] = r.Form
		api.requestForm["form"] = r.PostForm

		for k, v := range matchedController {
			controller := api.controllerRegistry[k]
			fmt.Println(controller, api.requestForm)

			in := make([]reflect.Value, 2)
			in[0] = reflect.ValueOf(api.requestForm)
			in[1] = reflect.ValueOf(r.Header)

			returnValues := reflect.ValueOf(controller).MethodByName(v).Call(in)

			statusCode := returnValues[0].Interface()
			statusCodeInt := statusCode.(int)

			response := returnValues[1].Interface()
			responseHeaders := http.Header{}

			if len(returnValues) == 3 {
				responseHeaders = returnValues[2].Interface().(http.Header)
			}

			api.JSONResponse(w, statusCodeInt, response, responseHeaders)
		}

		//fmt.Println( api.requestForm )
	}
}

func (api *APISerivce) JSONResponse(w http.ResponseWriter, code int, response interface{}, header http.Header) {

	for k, v := range header {
		w.Header().Add(k, v[0])
	}

	w.WriteHeader(code)

	resp, err := json.Marshal(response)

	if err != nil {
		fmt.Println("JSON err: ", err)
	}

	w.Write(resp)
}

func (api *APISerivce) RegisterHandleFunc() {

	fmt.Println("<=====")

	for k, v := range api.registeredPathAndController {

		fmt.Println(k, v)

		path := k

		if !strings.HasPrefix(path, "/") {
			path = fmt.Sprintf("/%s", path)
		}

		http.HandleFunc(path, api.HandleRequest(v))
	}

	fmt.Println("=======>")
}

func (api *APISerivce) Server(port int) {

	api.RegisterHandleFunc()

	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func NewService() *APISerivce {

	var apiService = new(APISerivce)

	apiService.controllerRegistry = make(map[string]interface{})
	apiService.registeredPathAndController = make(map[string]map[string]map[string]string)
	apiService.requestForm = make(map[string]url.Values)

	return apiService
}

type IndexController struct {
}

func (IndexController *IndexController) Index(request map[string]url.Values, headers http.Header) (statusCode int, response interface{}, responseHeader http.Header) {

	fmt.Println("IndexController@index: ", request, headers)

	return 200, map[string]string{"hello": "world" }, http.Header{"Foo": {"Bar"}}
}

func (IndexController *IndexController) Post() (statusCode int, response interface{}) {

	return 200, map[string]string{"hello": "post"}
}

func main() {

	apiService := NewService()

	apiService.Get("index", "IndexController@Index")
	apiService.Post("index", "IndexController@Post")
	apiService.Post("user", "UserController@Add")

	fmt.Println("=============\n")

	apiService.registerController(&IndexController{})

	fmt.Println(apiService.registeredPathAndController)

	apiService.Server(8899)
}
