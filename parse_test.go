package tf_docs

import (
	"fmt"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/token"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFindAndParse(t *testing.T) {
	cases := []struct {
		InputDir string
		Result   []*TFModule
	}{
		{
			InputDir: "./testdata/modules/depth1",
			Result: []*TFModule{
				{
					Title: "depth1",
					Link: "depth1",
					Variables: []*Variable{
						{
							Name:        "test",
							Type:        "string",
							Description: "this is a variable",
							Default:     "",
							Required:    true,
						},
					},
					Outputs: []*Output{
						{
							Description: "output description",
							Name:        "test",
						},
					},
					Resources: []*Resource{
						{
							Type:        "aws_ami",
							Name:        "ami",
							Description: "this is a resource",
						},
					},
					Modules:     []*Module{
						{
							Name:        "test",
							Description: "here's a module",
							Source:      "../",
						},
					},
					Description: "depth1 is a test module",
				},
			},
		},
		{
			InputDir: "./testdata/modules/depth2",
			Result: []*TFModule{
				{
					Title: "module1",
					Link: "module1",
					Variables: []*Variable{
						{
							Name:        "test",
							Type:        "string",
							Description: "this is a variable",
							Default:     "",
							Required:    true,
						},
					},
					Outputs: []*Output{
						{
							Description: "output description",
							Name:        "test",
						},
					},
					Resources: []*Resource{
						{
							Type:        "aws_ami",
							Name:        "ami",
							Description: "this is a resource",
						},
					},
					Modules:     []*Module{
						{
							Name:        "test",
							Description: "here's a module",
							Source:      "../",
						},
					},
					Description: "module1 is a test module",
				},
				{
					Title: "module2",
					Link: "module2",
					Variables: []*Variable{
						{
							Name:        "test",
							Type:        "string",
							Description: "this is a variable",
							Default:     "",
							Required:    true,
						},
					},
					Outputs: []*Output{
						{
							Description: "output description",
							Name:        "test",
						},
					},
					Resources: []*Resource{
						{
							Type:        "aws_ami",
							Name:        "ami",
							Description: "this is a resource",
						},
					},
					Modules:     []*Module{
						{
							Name:        "test",
							Description: "here's a module",
							Source:      "../",
						},
					},
					Description: "module2 is a test module",
				},
			},
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("LoadModuleFiles %v", i), func(t *testing.T) {
			result, err := FindAndParse(c.InputDir)
			assert.NoError(t, err, "")
			assert.Equal(t, c.Result, result, "")
		})
	}
}

func TestListModuleFiles(t *testing.T) {
	cases := []struct {
		DirPath string
		TFFiles []string
	}{
		{
			DirPath: "./testdata/modules/depth1",
			TFFiles: []string{
				"file.tf",
				"main.tf",
			},
		},
		{
			DirPath: "./testdata/modules/depth2/module1/",
			TFFiles: []string{
				"main.tf",
			},
		},
		{
			DirPath: "./testdata/modules/none",
			TFFiles: []string{},
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("listModuleFiles %v", i), func(t *testing.T) {
			result, err := ListModuleFiles(c.DirPath)
			assert.NoError(t, err, "Expected no error")
			assert.Equal(t, c.TFFiles, result, "")
		})
	}
}

func TestParse(t *testing.T) {
	cases := []struct {
		Input      []string
		ModuleName string
		Module     *TFModule
		Err        bool
	}{
		{
			Input: []string{
				`// test
variable "test" {
  type = "string"
  description = "desc"
}

// module that does a thing
module "test" {
  source = "../../test"
}

// resource desc
resource "aws" "test" {
  k = "v"
}

output "val" {
  value = "val"
  description = "output desc"
}
`,
`variable "test2" {
  type = "string"
  description = "desc"
}

// module that does a thing
module "test2" {
  source = "../../test"
}

// resource desc
resource "aws" "test2" {
  k = "v"
}

output "val2" {
  value = "val"
  description = "output desc"
}
`,
			},
			ModuleName: "test",
			Module: &TFModule{
				Title: "test",
				Variables: []*Variable{
					{
						Name:        "test",
						Type:        "string",
						Description: "desc",
						Default:     "",
						Required:    true,
					},
					{
						Name:        "test2",
						Type:        "string",
						Description: "desc",
						Default:     "",
						Required:    true,
					},
				},
				Outputs: []*Output{
					{
						Description: "output desc",
						Name:        "val",
					},
					{
						Description: "output desc",
						Name:        "val2",
					},
				},
				Resources: []*Resource{
					{
						Type:        "aws",
						Name:        "test",
						Description: "resource desc",
					},
					{
						Type:        "aws",
						Name:        "test2",
						Description: "resource desc",
					},
				},
				Modules: []*Module{
					{
						Name:        "test",
						Description: "module that does a thing",
						Source:      "../../test",
					},
					{
						Name:        "test2",
						Description: "module that does a thing",
						Source:      "../../test",
					},
				},
				Description: "test",
			},
			Err: false,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Parse %v", i), func(t *testing.T) {
			result, err := Parse(c.Input, c.ModuleName)
			if c.Err {
				assert.Error(t, err, "Expected an Error")
			} else {
				assert.NoError(t, err, "Expected no error")
				assert.Equal(t, c.Module, result, "expected ")
			}
		})
	}
}

func TestTraverseDirectories(t *testing.T) {
	cases := []struct {
		DirPath     string
		ModulePaths []string
	}{
		{
			DirPath:     "./testdata/modules/depth1",
			ModulePaths: []string{"./testdata/modules/depth1"},
		},
		{
			DirPath:     "./testdata/modules/depth2",
			ModulePaths: []string{"./testdata/modules/depth2/module1", "./testdata/modules/depth2/module2"},
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("traverseDirectories %v", i), func(t *testing.T) {
			result, err := traverseDirectory(c.DirPath)
			assert.NoError(t, err)
			assert.EqualValues(t, c.ModulePaths, result, "Should be equal")
		})
	}
}

func TestExtractDescription(t *testing.T) {
	cases := []struct {
		Comments   []*Comment
		ModuleName string
		Result     string
	}{
		{
			Comments: []*Comment{
				{
					Text: "test module description",
					Col:  0,
					Line: 1,
				},
				{
					Text: "test another module description",
					Col:  0,
					Line: 1,
				},
			},
			ModuleName: "test",
			Result:     "test module description",
		},
		{
			Comments: []*Comment{
				{
					Text: "test module description",
					Col:  0,
					Line: 2,
				},
				{
					Text: "test another module description",
					Col:  0,
					Line: 1,
				},
			},
			ModuleName: "test",
			Result:     "test another module description",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("extractDescription %v", i), func(t *testing.T) {
			result := extractDescription(c.Comments, c.ModuleName)
			assert.Equal(t, c.Result, result, "should be equal")
		})
	}
}

func TestExtractModules(t *testing.T) {
	cases := []struct {
		Input  []*Value
		Result []*Module
		Err    bool
	}{
		{
			Input: []*Value{
				{
					Key: map[string][]string{
						"module": {"name"},
					},
					Val: map[string]string{
						"source": "../../module",
					},
					Comment: Comment{
						Text: "describe a module",
						Col:  0,
						Line: 0,
					},
				},
			},
			Result: []*Module{
				{
					Name:        "name",
					Description: "describe a module",
					Source:      "../../module",
				},
			},
			Err: false,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("extractModules %v", i), func(t *testing.T) {
			result, err := extractModules(c.Input)
			if c.Err {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, c.Result, result, "expected equality")
			}
		})
	}
}

func TestExtractResources(t *testing.T) {
	cases := []struct {
		Input  []*Value
		Result []*Resource
		Err    bool
	}{
		{
			Input: []*Value{
				{
					Key: map[string][]string{
						"resource": {"type", "name"},
					},
					Val: map[string]string{},
					Comment: Comment{
						Text: "Describe me a resource",
						Col:  0,
						Line: 0,
					},
				},
			},
			Result: []*Resource{
				{
					Type:        "type",
					Name:        "name",
					Description: "Describe me a resource",
				},
			},
			Err: false,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("extractResources %v", i), func(t *testing.T) {
			result, err := extractResources(c.Input)
			if c.Err {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, c.Result, result, "expected equality")
			}
		})
	}
}

func TestExtractOutputs(t *testing.T) {
	cases := []struct {
		Input  []*Value
		Result []*Output
		Err    bool
	}{
		{
			Input: []*Value{
				{
					Key: map[string][]string{
						"output": {"hello"},
					},
					Val: map[string]string{
						"description": "An output shows things",
					},
					Comment: Comment{},
				},
				{
					Key: map[string][]string{
						"variable": {"hello"},
					},
					Val:     nil,
					Comment: Comment{},
				},
				{
					Key: map[string][]string{
						"output": {"world"},
					},
					Val: map[string]string{
						"description": "An output shows things",
					},
					Comment: Comment{},
				},
			},
			Result: []*Output{
				{
					Description: "An output shows things",
					Name:        "hello",
				},
				{
					Description: "An output shows things",
					Name:        "world",
				},
			},
			Err: false,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("extractVariables %v", i), func(t *testing.T) {
			result, err := extractOutputs(c.Input)
			if c.Err {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, c.Result, result, "expected equality")
			}
		})
	}
}

func TestExtractVariables(t *testing.T) {
	cases := []struct {
		Input  []*Value
		Result []*Variable
		Err    bool
	}{
		{
			Input: []*Value{
				{
					Key: map[string][]string{
						"variable": {"bah_humbug"},
					},
					Val: map[string]string{
						"type": "string",
					},
					Comment: Comment{},
				},
				{
					Key: map[string][]string{
						"variable": {"testing"},
					},
					Val: map[string]string{
						"type":        "string",
						"description": "a variable",
						"default":     "yes",
					},
					Comment: Comment{},
				},
				{
					Key: map[string][]string{
						"output": {"testing"},
					},
					Val: map[string]string{
						"type":  "string",
						"value": "yes",
					},
					Comment: Comment{},
				},
			},
			Result: []*Variable{
				{
					Name:        "bah_humbug",
					Type:        "string",
					Description: "",
					Default:     "",
					Required:    true,
				},
				{
					Name:        "testing",
					Type:        "string",
					Description: "a variable",
					Default:     "yes",
					Required:    false,
				},
			},
			Err: false,
		},
		{
			Input: []*Value{
				{
					Key: map[string][]string{
						"variable": {"bah_humbug"},
					},
					Val: map[string]string{
						"description": "no type",
					},
					Comment: Comment{},
				},
			},
			Result: []*Variable{},
			Err:    true,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("extractVariables %v", i), func(t *testing.T) {
			result, err := extractVariables(c.Input)
			if c.Err {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, c.Result, result, "expected equality")
			}
		})
	}
}

func TestExtractElements(t *testing.T) {
	cases := []struct {
		Type   string
		Input  []*Value
		Result []*Value
	}{
		{
			Type: "resource",
			Input: []*Value{
				{
					Key: map[string][]string{
						"resource": {"aws", "something"},
					},
					Val: map[string]string{
						"key": "value",
					},
					Comment: Comment{},
				},
				{
					Key: map[string][]string{
						"output": {"something"},
					},
					Val: map[string]string{
						"key": "value",
					},
					Comment: Comment{},
				},
			},
			Result: []*Value{
				{
					Key: map[string][]string{
						"resource": {"aws", "something"},
					},
					Val: map[string]string{
						"key": "value",
					},
					Comment: Comment{},
				},
			},
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("extractElements %v", i), func(t *testing.T) {
			result := extractElement(c.Input, c.Type)
			assert.Equal(t, c.Result, result, "should be equal")
		})
	}
}

func TestExtractValues(t *testing.T) {
	cases := []struct {
		Input  *ast.ObjectList
		Result []*Value
	}{
		{
			Input: &ast.ObjectList{
				Items: []*ast.ObjectItem{
					{
						Keys: []*ast.ObjectKey{
							{
								Token: token.Token{
									Type: 0,
									Pos:  token.Pos{},
									Text: "output",
									JSON: false,
								},
							},
							{
								Token: token.Token{
									Type: 0,
									Pos:  token.Pos{},
									Text: "test",
									JSON: false,
								},
							},
						},
						Assign: token.Pos{},
						Val: &ast.ObjectType{
							Lbrace: token.Pos{},
							Rbrace: token.Pos{},
							List: &ast.ObjectList{
								Items: []*ast.ObjectItem{
									{
										Keys: []*ast.ObjectKey{
											{
												Token: token.Token{
													Type: 0,
													Pos:  token.Pos{},
													Text: "value",
													JSON: false,
												},
											},
										},
										Assign: token.Pos{},
										Val: &ast.LiteralType{
											Token: token.Token{
												Type: 0,
												Pos:  token.Pos{},
												Text: "testVal",
												JSON: false,
											},
											LeadComment: nil,
											LineComment: nil,
										},
										LeadComment: nil,
										LineComment: nil,
									},
								},
							},
						},
						LeadComment: nil,
						LineComment: nil,
					},
				},
			},
			Result: []*Value{
				{
					Key: map[string][]string{
						"output": {"test"},
					},
					Val: map[string]string{
						"value": "testVal",
					},
					Comment: Comment{},
				},
			},
		},
		{
			Input: &ast.ObjectList{
				Items: []*ast.ObjectItem{
					{
						Keys: []*ast.ObjectKey{
							{
								Token: token.Token{
									Type: 0,
									Pos:  token.Pos{},
									Text: "resource",
									JSON: false,
								},
							},
							{
								Token: token.Token{
									Type: 0,
									Pos:  token.Pos{},
									Text: "aws_",
									JSON: false,
								},
							},
							{
								Token: token.Token{
									Type: 0,
									Pos:  token.Pos{},
									Text: "name",
									JSON: false,
								},
							},
						},
						Assign: token.Pos{},
						Val: &ast.ObjectType{
							Lbrace: token.Pos{},
							Rbrace: token.Pos{},
							List: &ast.ObjectList{
								Items: []*ast.ObjectItem{
									{
										Keys: []*ast.ObjectKey{
											{
												Token: token.Token{
													Type: 0,
													Pos:  token.Pos{},
													Text: "value",
													JSON: false,
												},
											},
										},
										Assign: token.Pos{},
										Val: &ast.LiteralType{
											Token: token.Token{
												Type: 0,
												Pos:  token.Pos{},
												Text: "testVal",
												JSON: false,
											},
											LeadComment: nil,
											LineComment: nil,
										},
										LeadComment: nil,
										LineComment: nil,
									},
								},
							},
						},
						LeadComment: nil,
						LineComment: nil,
					},
				},
			},
			Result: []*Value{
				{
					Key: map[string][]string{
						"resource": {"aws_", "name"},
					},
					Val: map[string]string{
						"value": "testVal",
					},
					Comment: Comment{},
				},
			},
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("extractValues %v", i), func(t *testing.T) {
			result := extractValues(c.Input)
			assert.Equal(t, c.Result, result, "should be equal")
		})
	}
}

func TestParseKeys(t *testing.T) {
	cases := []struct {
		Input  []*ast.ObjectKey
		Result map[string][]string
	}{
		{
			Input: []*ast.ObjectKey{
				{
					Token: token.Token{
						Type: 0,
						Pos:  token.Pos{},
						Text: "hello",
						JSON: false,
					},
				},
				{
					Token: token.Token{
						Type: 0,
						Pos:  token.Pos{},
						Text: "one",
						JSON: false,
					},
				},
				{
					Token: token.Token{
						Type: 0,
						Pos:  token.Pos{},
						Text: "two",
						JSON: false,
					},
				},
			},
			Result: map[string][]string{
				"hello": {"one", "two"},
			},
		},
		{
			Input: []*ast.ObjectKey{
				{
					Token: token.Token{
						Type: 0,
						Pos:  token.Pos{},
						Text: "hello",
						JSON: false,
					},
				},
			},
			Result: map[string][]string{
				"hello": {},
			},
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("parseKeys %v", i), func(t *testing.T) {
			result := parseKeys(c.Input)
			assert.Equal(t, c.Result, result, "should be equal")
		})
	}
}

func TestParseValues(t *testing.T) {
	cases := []struct {
		Input  *ast.ObjectType
		Result map[string]string
	}{
		{
			Input: &ast.ObjectType{
				List: &ast.ObjectList{
					Items: []*ast.ObjectItem{
						{
							Keys: []*ast.ObjectKey{
								{
									Token: token.Token{
										Type: 0,
										Pos:  token.Pos{},
										Text: "hello",
										JSON: false,
									},
								},
							},
							Assign: token.Pos{},
							Val: &ast.LiteralType{
								Token: token.Token{
									Type: 0,
									Pos:  token.Pos{},
									Text: "world",
									JSON: false,
								},
								LeadComment: nil,
								LineComment: nil,
							},
							LeadComment: nil,
							LineComment: nil,
						},
						{
							Keys: []*ast.ObjectKey{
								{
									Token: token.Token{
										Type: 0,
										Pos:  token.Pos{},
										Text: "hi",
										JSON: false,
									},
								},
							},
							Assign: token.Pos{},
							Val: &ast.ListType{
								Lbrack: token.Pos{},
								Rbrack: token.Pos{},
								List: []ast.Node{
									&ast.LiteralType{
										Token: token.Token{
											Type: 0,
											Pos:  token.Pos{},
											Text: "world",
											JSON: false,
										},
										LeadComment: nil,
										LineComment: nil,
									},
									&ast.LiteralType{
										Token: token.Token{
											Type: 0,
											Pos:  token.Pos{},
											Text: "world2",
											JSON: false,
										},
										LeadComment: nil,
										LineComment: nil,
									},
								},
							},
							LeadComment: nil,
							LineComment: nil,
						},
					},
				},
			},
			Result: map[string]string{
				"hello": "world",
				"hi":    "[world, world2]",
			},
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("parseValues %v", i), func(t *testing.T) {
			result := parseValues(c.Input)
			assert.Equal(t, c.Result, result, "should be equal")
		})
	}
}

func TestExtractComments(t *testing.T) {
	cases := []struct {
		Input  []*ast.CommentGroup
		Result []*Comment
	}{
		{
			Input: []*ast.CommentGroup{
				{
					List: []*ast.Comment{
						{
							Start: token.Pos{},
							Text:  "// Hello",
						},
					},
				},
				{
					List: []*ast.Comment{
						{
							Start: token.Pos{},
							Text:  "// World",
						},
					},
				},
			},
			Result: []*Comment{
				{
					Text: "Hello",
					Col:  0,
					Line: 0,
				},
				{
					Text: "World",
					Col:  0,
					Line: 0,
				},
			},
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("extractComments %v", i), func(t *testing.T) {
			t.Log("input len", len(c.Input))
			result := extractComments(c.Input)
			assert.Equal(t, len(c.Result), len(result), "")
			assert.Equal(t, c.Result, result, "")
		})
	}
}

func TestTidyComment(t *testing.T) {
	cases := []struct {
		Input  string
		Result string
	}{
		{
			Input:  "// Here's a comment",
			Result: "Here's a comment",
		},
		{
			Input:  "//Another comment",
			Result: "Another comment",
		},
		{
			Input:  "// Trailing space   ",
			Result: "Trailing space",
		},
		{
			Input:  "//    Leading space",
			Result: "Leading space",
		},
		{
			Input:  "/* Multiline on one */",
			Result: "Multiline on one",
		},
		{
			Input: `/* True
multiline
comment */`,
			Result: "True\nmultiline\ncomment",
		},
		{
			Input: `/* True
multiline
comment 
*/`,
			Result: "True\nmultiline\ncomment",
		},
		{
			Input: `/* 
True
multiline
comment */`,
			Result: "True\nmultiline\ncomment",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("tidyComment %v", i), func(t *testing.T) {
			result := tidyComment(c.Input)
			assert.Equal(t, c.Result, result, "should be equal")
		})
	}
}

func TestParseComment(t *testing.T) {
	cases := []struct {
		Input   *ast.CommentGroup
		Comment Comment
		Err     bool
	}{
		{
			Input: &ast.CommentGroup{
				List: []*ast.Comment{
					{
						Start: token.Pos{},
						Text:  "// Hello",
					},
				},
			},
			Comment: Comment{
				Text: "Hello",
				Col:  0,
				Line: 0,
			},
			Err: false,
		},
		{
			Input: &ast.CommentGroup{
				List: []*ast.Comment{
					{
						Start: token.Pos{},
						Text:  "// Hello",
					},
					{
						Start: token.Pos{},
						Text:  "// World",
					},
				},
			},
			Comment: Comment{
				Text: "Hello World",
				Col:  0,
				Line: 0,
			},
			Err: false,
		},
		{
			Input:   nil,
			Comment: Comment{},
			Err:     true,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("parseCommment %v", i), func(t *testing.T) {
			result, err := parseComment(c.Input)
			if c.Err {
				assert.Errorf(t, err, "Expected no error")
			} else {
				assert.NoError(t, err, "Expected no error")
				assert.EqualValues(t, c.Comment, result, "Expected comment to be equal")
			}
		})
	}
}

func TestTrimStrings(t *testing.T) {
	cases := []struct {
		Input  string
		Result string
	}{
		{
			Input:  "\"hello\"",
			Result: "hello",
		},
		{
			Input:  "\\\"hello\\\"",
			Result: "hello",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("trimStrings %v", i), func(t *testing.T) {
			result := trimStrings(c.Input)
			assert.Equal(t, c.Result, result, "should be equal")
		})
	}
}
