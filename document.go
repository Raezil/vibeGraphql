package vibeGraphql

type Node interface {
	TokenLiteral() string
}

type Document struct {
	Definitions []Definition
}

func (d *Document) TokenLiteral() string {
	if len(d.Definitions) > 0 {
		return d.Definitions[0].TokenLiteral()
	}
	return ""
}

type Definition interface {
	Node
}

type OperationDefinition struct {
	Operation           string
	Name                string
	VariableDefinitions []VariableDefinition
	SelectionSet        *SelectionSet
}

func (op *OperationDefinition) TokenLiteral() string {
	if op.Name != "" {
		return op.Name
	}
	return op.Operation
}

type VariableDefinition struct {
	Variable string
	Type     Type
}

func (v *VariableDefinition) TokenLiteral() string {
	return v.Variable
}

type Type struct {
	Name    string
	NonNull bool
	IsList  bool  // Indicates if the type is a list
	Elem    *Type // The element type if IsList is true
}

type SelectionSet struct {
	Selections []Selection
}

type Selection interface {
	Node
}

type Field struct {
	Name         string
	Arguments    []Argument
	SelectionSet *SelectionSet
}

func (f *Field) TokenLiteral() string {
	return f.Name
}

type Argument struct {
	Name  string
	Value *Value
}

func (a *Argument) TokenLiteral() string {
	return a.Name
}

type Value struct {
	Kind         string // "Int", "String", "Boolean", "Variable", "Enum", "Object", "Array"
	Literal      string
	ObjectFields map[string]*Value // for nested object values (if Kind == "Object")
	List         []*Value          // for array values (if Kind == "Array")
}

func (v *Value) TokenLiteral() string {
	return v.Literal
}
