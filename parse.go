package tf_docs

import (
	"fmt"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"io/ioutil"
	"strings"
)

type TFModule struct {
	Path        string
	Title       string
	Link        string
	Variables   []*Variable
	Outputs     []*Output
	Resources   []*Resource
	Modules     []*Module
	Description string
}

type Comment struct {
	Text string
	Col  int
	Line int
}

type Variable struct {
	Name        string
	Type        string
	Description string
	Default     string
	Required    bool
}

type Output struct {
	Description string
	Name        string
}

type Resource struct {
	Type        string
	Name        string
	Description string
}

type Module struct {
	Name        string
	Description string
	Source      string
}

type Value struct {
	Key     map[string][]string
	Val     map[string]string
	Comment Comment
}

const (
	VARIABLE = "variable"
	OUTPUT   = "output"
	MODULE   = "module"
	RESOURCE = "resource"
)

// FindAndParse finds all of the modules within a directory and parses them all.
func FindAndParse(directory string) ([]*TFModule, error) {
	var modules []*TFModule

	directoryDepth := len(strings.Split(directory, "/"))

	if directory == "" {
		return modules, fmt.Errorf("directory cannot be empty")
	}

	modulesDirs, err := traverseDirectory(directory)
	if err != nil {
		return modules, err
	}
	if len(modulesDirs) == 0 {
		return modules, fmt.Errorf("no modules found in path %s", directory)
	}

	for _, d := range modulesDirs {
		var moduleFiles []string
		files, err := ListModuleFiles(d)
		if err != nil {
			return modules, err
		}
		for _, file := range files {
			fileBody, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", d, file))
			if err != nil {
				return modules, err
			}
			moduleFiles = append(moduleFiles, string(fileBody))
		}
		splitModule := strings.Split(d, "/")
		moduleName := splitModule[len(splitModule)-1]
		tfFile, err := Parse(moduleFiles, moduleName)
		if err != nil {
			return modules, err
		}
		if len(splitModule) > directoryDepth + 1 {
			tfFile.Path = strings.Join(splitModule[directoryDepth:len(splitModule)-1], "/")
		}
		tfFile.Link = strings.Replace(tfFile.Path, "/", "-", -1) + "_" + tfFile.Title
		if strings.HasPrefix(tfFile.Link, "_") {
			tfFile.Link = strings.Replace(tfFile.Link, "_", "", 1)
		}
		modules = append(modules, tfFile)
	}

	return modules, nil
}

// listModuleFiles returns a list ouf files with a .tf extension
// within a directory.
func ListModuleFiles(directory string) ([]string, error) {
	files := []string{}

	fileInfos, err := ioutil.ReadDir(directory)
	if err != nil {
		return files, err
	}
	for _, i := range fileInfos {
		if strings.HasSuffix(i.Name(), ".tf") {
			files = append(files, i.Name())
		}
	}

	return files, nil
}

// traverseDirectory traverses directories and returns a list of directories that contain .tf files.
func traverseDirectory(directory string) ([]string, error) {
	var directoryPaths []string

	fileInfos, err := ioutil.ReadDir(directory)
	if err != nil {
		return directoryPaths, err
	}

	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".tf") {
			directoryPaths = append(directoryPaths, directory)
			break
		}
		if fileInfo.IsDir() {
			directories, err := traverseDirectory(fmt.Sprintf("%s/%s", directory, fileInfo.Name()))
			if err != nil {
				return directoryPaths, nil
			}
			directoryPaths = append(directoryPaths, directories...)
		}
	}

	return directoryPaths, nil
}

// Parse generates a TFModule given a number of Terraform files (as strings) as input
func Parse(hclText []string, moduleName string) (*TFModule, error) {
	result := &TFModule{}

	if moduleName == "" {
		return result, fmt.Errorf("error moduleName cannot be empty")
	}

	result.Title = moduleName

	var hclParseTrees []*ast.File

	for _, text := range hclText {
		hclParseTree, err := hcl.Parse(text)
		if err != nil {
			return nil, err
		}
		hclParseTrees = append(hclParseTrees, hclParseTree)
	}

	var comments []*Comment
	var variables []*Variable
	var outputs []*Output
	var modules []*Module
	var resources []*Resource

	for _, hclTree := range hclParseTrees {
		comments = append(comments, extractComments(hclTree.Comments)...)
		values := extractValues(hclTree.Node)

		tmpVariables, err := extractVariables(values)
		if err != nil {
			return result, err
		}
		variables = append(variables, tmpVariables...)

		tmpOutputs, err := extractOutputs(values)
		if err != nil {
			return result, err
		}
		outputs = append(outputs, tmpOutputs...)

		tmpModules, err := extractModules(values)
		if err != nil {
			return result, err
		}
		modules = append(modules, tmpModules...)

		tmpResources, err := extractResources(values)
		if err != nil {
			return result, err
		}
		resources = append(resources, tmpResources...)
	}
	description := extractDescription(comments, moduleName)

	result.Variables = variables
	result.Outputs = outputs
	result.Description = description
	result.Modules = modules
	result.Resources = resources

	return result, nil
}

// extractDescription parse each comment and returns the first comment that starts with the moduleName
// to use as the overall description of the module.
func extractDescription(comments []*Comment, moduleName string) string {
	var description string

	for _, comment := range comments {
		if comment.Line != 1 {
			continue
		}
		commentText := strings.TrimSpace(comment.Text)
		if strings.HasPrefix(commentText, moduleName) {
			description = commentText
			break
		}
	}

	return description
}

// extractModule iterates over each Value, selects those that are modules and returns a slice of
// Modules generated from the matching values. Checks that the module is names and has a source.
func extractModules(values []*Value) ([]*Module, error) {
	var modules []*Module

	mods := extractElement(values, MODULE)
	for _, m := range mods {
		if _, ok := m.Key["module"]; !ok {
			return modules, fmt.Errorf("name is required for a module")
		}
		if _, ok := m.Val["source"]; !ok {
			return modules, fmt.Errorf("source is required for a variable")
		}

		module := &Module{}
		name := m.Key["module"]
		module.Name = name[0]
		module.Description = m.Comment.Text
		module.Source = m.Val["source"]

		modules = append(modules, module)
	}

	return modules, nil
}

// extractResources iterates over each value, selects those that are resources and returns a slice of
// Resources generated from those matching values. Checks that the Resource has a name.
func extractResources(values []*Value) ([]*Resource, error) {
	var resources []*Resource

	mods := extractElement(values, RESOURCE)
	for _, m := range mods {
		if _, ok := m.Key["resource"]; !ok {
			return resources, fmt.Errorf("name is required for a resource")
		}

		resource := &Resource{}
		keys := m.Key["resource"]
		resource.Name = keys[1]
		resource.Type = keys[0]
		resource.Description = m.Comment.Text

		resources = append(resources, resource)
	}

	return resources, nil
}

// extractOutputs iterates over each value, selects those that are outputs and returns a slice of
// Outputs generated from those matching values. Checks that the Output has a name.
func extractOutputs(values []*Value) ([]*Output, error) {
	var outputs []*Output

	outs := extractElement(values, OUTPUT)
	for _, o := range outs {
		if _, ok := o.Key["output"]; !ok {
			return outputs, fmt.Errorf("name is required for a output")
		}
		output := &Output{}

		name := o.Key["output"]
		output.Name = name[0]
		if _, ok := o.Val["description"]; ok {
			output.Description = o.Val["description"]
		}

		outputs = append(outputs, output)
	}

	return outputs, nil
}

// extractVariables iterates over each value, selects those that are variables and returns a slice of
// Variables generated from those matching values. Checks that the Variable has a name and a type.
func extractVariables(values []*Value) ([]*Variable, error) {
	var variables []*Variable

	vars := extractElement(values, VARIABLE)
	for _, v := range vars {
		if _, ok := v.Key["variable"]; !ok {
			return variables, fmt.Errorf("name is required for a variable")
		}
		if _, ok := v.Val["type"]; !ok {
			return variables, fmt.Errorf("type is required for a variable")
		}

		variable := &Variable{}

		name := v.Key["variable"]
		variable.Name = name[0]
		variable.Type = v.Val["type"]
		if _, ok := v.Val["description"]; ok {
			variable.Description = v.Val["description"]
		}
		if _, ok := v.Val["default"]; ok {
			variable.Default = v.Val["default"]
		}
		if variable.Default == "" {
			variable.Required = true
		}

		variables = append(variables, variable)
	}

	return variables, nil
}

// extractElement returns all the Values, from a list of Values, that have a specified key.
func extractElement(values []*Value, elementType string) []*Value {
	var result []*Value

	for _, value := range values {
		if _, ok := value.Key[elementType]; ok {
			result = append(result, value)
		}
	}

	return result
}

// extractValues returns all of the Values from an ast.Node
func extractValues(node ast.Node) []*Value {
	var values []*Value

	items := node.(*ast.ObjectList).Items

	for _, item := range items {
		value := &Value{}
		switch item.Val.(type) {
		case *ast.ObjectType:
			val := parseValues(item.Val.(*ast.ObjectType))
			key := parseKeys(item.Keys)
			comment, _ := parseComment(item.LeadComment)

			value.Val = val
			value.Key = key
			value.Comment = comment
		}

		values = append(values, value)
	}

	return values
}

// parseKeys returns the first key from a slice of *ast.ObjectKey mapped to a slice of strings consisting
// of values for that key
func parseKeys(rawKeys []*ast.ObjectKey) map[string][]string {
	result := map[string][]string{}

	key := rawKeys[0].Token.Text
	values := []string{}
	for i := 1; i < len(rawKeys); i++ {
		values = append(values, trimStrings(rawKeys[i].Token.Text))
	}

	result[key] = values

	return result
}

func parseValues(rawValue *ast.ObjectType) map[string]string {
	result := map[string]string{}

	for _, item := range rawValue.List.Items {
		switch item.Val.(type) {
		case *ast.LiteralType:
			result[trimStrings(item.Keys[0].Token.Text)] = trimStrings(item.Val.(*ast.LiteralType).Token.Text)
		case *ast.ListType:
			var valueList []string
			for _, v := range item.Val.(*ast.ListType).List {
				switch v.(type) {
				case *ast.LiteralType:
					valueList = append(valueList, trimStrings(v.(*ast.LiteralType).Token.Text))
				}
			}
			joinedList := strings.Join(valueList, ", ")
			result[trimStrings(item.Keys[0].Token.Text)] = fmt.Sprintf("[%s]", joinedList)
		default:
		}
	}

	return result
}

func extractComments(commentGroup []*ast.CommentGroup) []*Comment {
	var comments []*Comment

	rawComments := commentGroup
	for _, rawComment := range rawComments {
		comment, err := parseComment(rawComment)
		if err != nil {
			fmt.Println("eerrrr", err)
			continue
		}
		comments = append(comments, &comment)
	}
	return comments
}

func tidyComment(comment string) string {
	result := comment
	if strings.HasPrefix(comment, "//") {
		result = strings.TrimPrefix(result, "//")
	}
	if strings.HasPrefix(comment, "/*") {
		result = strings.TrimPrefix(result, "/*")
		result = strings.TrimSuffix(result, "*/")
	}
	result = strings.TrimSpace(result)
	return result
}

func parseComment(rawComment *ast.CommentGroup) (Comment, error) {
	var comment Comment
	if rawComment == nil {
		return comment, fmt.Errorf("comment is nil")
	}

	var comments []string
	for _, c := range rawComment.List {
		commentText := tidyComment(c.Text)
		comments = append(comments, commentText)
	}

	commentString := strings.Join(comments, " ")

	comment.Text = commentString
	comment.Col = rawComment.Pos().Column
	comment.Line = rawComment.Pos().Line

	return comment, nil
}

func trimStrings(input string) string {
	trimPatterns := []string{"\\\"", "\""}
	result := input

	for _, pattern := range trimPatterns {
		result = strings.TrimPrefix(result, pattern)
		result = strings.TrimSuffix(result, pattern)
	}
	return result
}
