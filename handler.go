package vibeGraphql

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

func resolveArgument(arg *Argument, variables map[string]interface{}) (interface{}, error) {
	switch arg.Value.Kind {
	case "Int":
		return strconv.Atoi(arg.Value.Literal)
	case "String":
		return arg.Value.Literal, nil
	case "Boolean":
		return arg.Value.Literal == "true", nil
	case "Variable":
		if val, ok := variables[arg.Value.Literal]; ok {
			return val, nil
		}
		return nil, fmt.Errorf("variable %s not provided", arg.Value.Literal)
	default:
		return arg.Value.Literal, nil
	}
}

// executeDocument processes the parsed AST and returns a response.
func executeDocument(doc *Document, variables map[string]interface{}) (map[string]interface{}, error) {
	response := map[string]interface{}{}
	// For simplicity, we assume one operation definition.
	if len(doc.Definitions) == 0 {
		return response, fmt.Errorf("no definitions found")
	}
	op, ok := doc.Definitions[0].(*OperationDefinition)
	if !ok {
		return response, fmt.Errorf("unsupported definition type")
	}
	// Execute the top-level selection set (root query)
	data, err := executeSelectionSet(nil, op.SelectionSet, variables)
	if err != nil {
		return response, err
	}
	response["data"] = data
	return response, nil
}

// resolveField looks up the appropriate resolver for a field. When the source is nil (top-level),
// it checks both QueryResolvers and MutationResolvers. For nested fields, it falls back to reflective
// lookup on the source object.
func resolveField(source interface{}, field *Field, variables map[string]interface{}) (interface{}, error) {
	// At the top level, source is nil, so try both query and mutation resolvers.
	if source == nil {
		// First, try the query resolver.
		if resolver, ok := QueryResolvers[field.Name]; ok {
			args := buildArgs(field, variables)
			return resolver(source, args)
		}
		// Next, try the mutation resolver.
		if resolver, ok := MutationResolvers[field.Name]; ok {
			args := buildArgs(field, variables)
			return resolver(source, args)
		}
	}

	// If the source is not nil, or no top-level resolver is found,
	// fallback to reflective lookup on the source (if it's a struct).
	// (This is optional; you may want to require resolvers for all top-level fields.)
	if source != nil {
		// Use reflection or your existing logic to resolve nested fields.
		// For brevity, we'll assume the reflective resolution is implemented elsewhere.
		return reflectResolve(source, field)
	}

	return nil, fmt.Errorf("no resolver found for field %s", field.Name)
}

// reflectResolve is a helper that uses reflection to find a field value
// on a source struct. (Implementation not shown here.)
func reflectResolve(source interface{}, field *Field) (interface{}, error) {
	val := reflect.ValueOf(source)
	// Dereference pointer if needed.
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, fmt.Errorf("source is nil")
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("source is not a struct")
	}

	typ := val.Type()
	// Loop through all fields.
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		// Check if the field name matches (case-insensitive).
		if strings.EqualFold(sf.Name, field.Name) {
			return val.Field(i).Interface(), nil
		}
		// Also check the "json" tag if present.
		if tag, ok := sf.Tag.Lookup("json"); ok {
			// The tag may contain options like "id,omitempty"; split them.
			tagName := strings.Split(tag, ",")[0]
			if strings.EqualFold(tagName, field.Name) {
				return val.Field(i).Interface(), nil
			}
		}
	}

	return nil, fmt.Errorf("no resolver found for field %s via reflection", field.Name)
}

// buildArgs constructs a map of argument names to values extracted
// from field.Arguments. If an argument is a variable, its value is looked
// up in the provided variables map.
// buildArgs constructs a map of argument names to Go values.
// It recursively handles nested object arguments.
func buildArgs(field *Field, variables map[string]interface{}) map[string]interface{} {
	args := make(map[string]interface{})
	for _, arg := range field.Arguments {
		args[arg.Name] = buildValue(arg.Value, variables)
	}
	return args
}

// buildValue converts a Value to a corresponding Go value.
// It handles variables, basic scalar types, and nested object values.
func buildValue(val *Value, variables map[string]interface{}) interface{} {
	switch val.Kind {
	case "Variable":
		if v, ok := variables[val.Literal]; ok {
			return v
		}
		return nil
	case "Int":
		i, err := strconv.Atoi(val.Literal)
		if err != nil {
			return 0
		}
		return i
	case "String":
		return val.Literal
	case "Boolean":
		return val.Literal == "true"
	case "Object":
		m := make(map[string]interface{})
		for key, fieldVal := range val.ObjectFields {
			m[key] = buildValue(fieldVal, variables)
		}
		return m
	case "Array":
		arr := []interface{}{}
		for _, elem := range val.List {
			arr = append(arr, buildValue(elem, variables))
		}
		return arr
	default:
		return val.Literal
	}
}

// executeSelectionSet traverses the selection set, resolves each field,
// and uses resolveNestedSelection to process any nested selections.
func executeSelectionSet(source interface{}, ss *SelectionSet, variables map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, sel := range ss.Selections {
		field, ok := sel.(*Field)
		if !ok {
			continue
		}
		// Resolve the field based on the current source.
		res, err := resolveField(source, field, variables)
		if err != nil {
			return nil, err
		}
		// If the field has nested selections, process them.
		if field.SelectionSet != nil {
			nested, err := resolveNestedSelection(res, field.SelectionSet, variables)
			if err != nil {
				return nil, err
			}
			result[field.Name] = nested
		} else {
			result[field.Name] = res
		}
	}
	return result, nil
}

// resolveNestedSelection handles nested selection sets by examining the
// resolved value. It supports both single objects (e.g. *User) and slices (e.g. []*User).
func resolveNestedSelection(res interface{}, ss *SelectionSet, variables map[string]interface{}) (interface{}, error) {
	val := reflect.ValueOf(res)
	switch val.Kind() {
	case reflect.Ptr:
		// If pointer is nil, return as is.
		if val.IsNil() {
			return res, nil
		}
		// If pointer to struct, process the struct.
		if val.Elem().Kind() == reflect.Struct {
			return executeSelectionSet(res, ss, variables)
		}
	case reflect.Struct:
		return executeSelectionSet(res, ss, variables)
	case reflect.Slice:
		var arr []interface{}
		for i := 0; i < val.Len(); i++ {
			item := val.Index(i).Interface()
			sub, err := executeSelectionSet(item, ss, variables)
			if err != nil {
				return nil, err
			}
			arr = append(arr, sub)
		}
		return arr, nil
	}
	// For other scalar types or unsupported kinds, return the original value.
	return res, nil
}

func GraphqlHandler(w http.ResponseWriter, r *http.Request) {
	// Expect a JSON body with at least a "query" field.
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Variables == nil {
		req.Variables = make(map[string]interface{})
	}

	// Lex and parse the query.
	lexer := NewLexer(req.Query)
	parser := NewParser(lexer)
	doc := parser.ParseDocument()

	// Execute the query.
	result, err := executeDocument(doc, req.Variables)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the JSON result.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// executeSubscription calls the registered subscription resolver and returns a channel.
// The resolver should return a channel (i.e. <-chan interface{}) with subscription events.
// executeSubscription calls the registered subscription resolver and returns a channel.
// The resolver should return either a chan interface{} or a <-chan interface{}.
func executeSubscription(source interface{}, field *Field, variables map[string]interface{}) (<-chan interface{}, error) {
	if resolver, ok := SubscriptionResolvers[field.Name]; ok {
		args := buildArgs(field, variables)
		res, err := resolver(source, args)
		if err != nil {
			return nil, err
		}
		// Try to type assert to a read-only channel.
		if ch, ok := res.(<-chan interface{}); ok {
			return ch, nil
		}
		// Otherwise, try to type assert to a bidirectional channel and convert it.
		if ch, ok := res.(chan interface{}); ok {
			return (<-chan interface{})(ch), nil
		}
		return nil, fmt.Errorf("subscription resolver for field %s did not return a channel", field.Name)
	}
	return nil, fmt.Errorf("no subscription resolver found for field %s", field.Name)
}

// upgrader upgrades HTTP connections to WebSocket connections.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// SubscriptionRequest represents the expected JSON payload for a subscription request.
type SubscriptionRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// SubscriptionHandler handles incoming subscription requests over WebSocket.
func SubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP to WebSocket.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Before upgrade, it's safe to use http.Error.
		http.Error(w, "unable to upgrade to websocket", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	// Read the subscription request from the WebSocket.
	_, msg, err := conn.ReadMessage()
	if err != nil {
		// After upgrade, write error messages directly to the WebSocket.
		conn.WriteMessage(websocket.TextMessage, []byte("failed to read subscription message"))
		return
	}

	var req SubscriptionRequest
	if err := json.Unmarshal(msg, &req); err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("invalid subscription JSON"))
		return
	}

	// Lex, parse, and extract the subscription operation.
	lexer := NewLexer(req.Query)
	parser := NewParser(lexer)
	doc := parser.ParseDocument()

	if len(doc.Definitions) == 0 {
		conn.WriteMessage(websocket.TextMessage, []byte("no subscription definition found"))
		return
	}

	op, ok := doc.Definitions[0].(*OperationDefinition)
	if !ok || op.Operation != "subscription" {
		conn.WriteMessage(websocket.TextMessage, []byte("provided operation is not a subscription"))
		return
	}

	if len(op.SelectionSet.Selections) == 0 {
		conn.WriteMessage(websocket.TextMessage, []byte("subscription selection set is empty"))
		return
	}

	field, ok := op.SelectionSet.Selections[0].(*Field)
	if !ok {
		conn.WriteMessage(websocket.TextMessage, []byte("invalid subscription field"))
		return
	}

	// Execute the subscription.
	subCh, err := executeSubscription(nil, field, req.Variables)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("subscription error: %v", err)))
		return
	}

	// Stream events from the subscription channel to the WebSocket.
	for event := range subCh {
		if err := conn.WriteJSON(event); err != nil {
			fmt.Printf("failed to write event: %v\n", err)
			break
		}
	}
}

// GraphqlUploadHandler supports both regular JSON GraphQL requests and multipart uploads.
// GraphqlUploadHandler handles multipart/form-data requests for file uploads.
func GraphqlUploadHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		GraphqlHandler(w, r)
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}
	operations := r.FormValue("operations")
	if operations == "" {
		http.Error(w, "missing operations field", http.StatusBadRequest)
		return
	}
	var req struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}
	if err := json.Unmarshal([]byte(operations), &req); err != nil {
		http.Error(w, "invalid operations JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Variables == nil {
		req.Variables = make(map[string]interface{})
	}
	fileMapStr := r.FormValue("map")
	if fileMapStr == "" {
		http.Error(w, "missing map field", http.StatusBadRequest)
		return
	}
	var fileMap map[string][]string
	if err := json.Unmarshal([]byte(fileMapStr), &fileMap); err != nil {
		http.Error(w, "invalid map JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	// Process each file and inject into variables.
	for fileKey, paths := range fileMap {
		file, header, err := r.FormFile(fileKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to retrieve file %s: %v", fileKey, err), http.StatusBadRequest)
			return
		}
		defer file.Close()
		fileData, err := ioutil.ReadAll(file)
		if err != nil {
			http.Error(w, "failed to read file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		for _, path := range paths {
			// If the path starts with "variables.", remove that prefix.
			adjustedPath := path
			if strings.HasPrefix(path, "variables.") {
				adjustedPath = strings.TrimPrefix(path, "variables.")
			}
			setNestedValue(req.Variables, adjustedPath, map[string]interface{}{
				"filename": header.Filename,
				"data":     fileData,
			})
		}
	}
	lexer := NewLexer(req.Query)
	parser := NewParser(lexer)
	doc := parser.ParseDocument()
	result, err := executeDocument(doc, req.Variables)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// setNestedValue assigns the given value into the nested map structure of vars based on the dot-separated path.
// For example, path "input.file" will set vars["input"]["file"] = value.
func setNestedValue(vars map[string]interface{}, path string, value interface{}) {
	keys := strings.Split(path, ".")
	current := vars
	for i, key := range keys {
		if i == len(keys)-1 {
			current[key] = value
		} else {
			if next, ok := current[key].(map[string]interface{}); ok {
				current = next
			} else {
				newMap := make(map[string]interface{})
				current[key] = newMap
				current = newMap
			}
		}
	}
}
