package treesitter

// NOTE: Go doesn't support multi-line raw strings quite like JS/TS template literals.
// It's often cleaner to load these from separate .scm files, but for direct translation,
// we'll use backticks. Be mindful of escaping if necessary.

const (
	// Combined JavaScript/JSON Query
	javascriptQuery = `
(
  (comment)* @doc
  .
  (method_definition
    name: (property_identifier) @name) @definition.method
  (#not-eq? @name "constructor")
  (#strip! @doc "^[\\s\\*/]+|^[\\s\\*/]$")
  (#select-adjacent! @doc @definition.method)
)

(
  (comment)* @doc
  .
  [
    (class
      name: (_) @name)
    (class_declaration
      name: (_) @name)
  ] @definition.class
  (#strip! @doc "^[\\s\\*/]+|^[\\s\\*/]$")
  (#select-adjacent! @doc @definition.class)
)

(
  (comment)* @doc
  .
  [
    (function_declaration
      name: (identifier) @name)
    (generator_function_declaration
      name: (identifier) @name)
  ] @definition.function
  (#strip! @doc "^[\\s\\*/]+|^[\\s\\*/]$")
  (#select-adjacent! @doc @definition.function)
)

(
  (comment)* @doc
  .
  (lexical_declaration
    (variable_declarator
      name: (identifier) @name
      value: [(arrow_function) (function_expression)]) @definition.function)
  (#strip! @doc "^[\\s\\*/]+|^[\\s\\*/]$")
  (#select-adjacent! @doc @definition.function)
)

(
  (comment)* @doc
  .
  (variable_declaration
    (variable_declarator
      name: (identifier) @name
      value: [(arrow_function) (function_expression)]) @definition.function)
  (#strip! @doc "^[\\s\\*/]+|^[\\s\\*/]$")
  (#select-adjacent! @doc @definition.function)
)

; JSON object definitions
(object) @object.definition

; JSON object key-value pairs
(pair
  key: (string) @property.name.definition
  value: [
    (object) @object.value
    (array) @array.value
    (string) @string.value
    (number) @number.value
    (true) @boolean.value
    (false) @boolean.value
    (null) @null.value
  ]
) @property.definition

; JSON array definitions
(array) @array.definition
`

	typescriptQuery = `

(function_signature
  name: (identifier) @name.definition.function) @definition.function

(method_signature
  name: (property_identifier) @name.definition.method) @definition.method

(abstract_method_signature
  name: (property_identifier) @name.definition.method) @definition.method

(abstract_class_declaration
  name: (type_identifier) @name.definition.class) @definition.class

(module
  name: (identifier) @name.definition.module) @definition.module

(function_declaration
  name: (identifier) @name.definition.function) @definition.function

(method_definition
  name: (property_identifier) @name.definition.method) @definition.method

(class_declaration
  name: (type_identifier) @name.definition.class) @definition.class

(call_expression
  function: (identifier) @func_name
  arguments: (arguments
    (string) @name
    [(arrow_function) (function_expression)]) @definition.test)
  (#match? @func_name "^(describe|test|it)$")

(assignment_expression
  left: (member_expression
    object: (identifier) @obj
    property: (property_identifier) @prop)
  right: [(arrow_function) (function_expression)]) @definition.test
  (#eq? @obj "exports")
  (#eq? @prop "test")
(arrow_function) @definition.lambda

; Switch statements and case clauses
(switch_statement) @definition.switch

; Individual case clauses with their blocks
(switch_case) @definition.case

; Default clause
(switch_default) @definition.default

; Enum declarations
(enum_declaration
  name: (identifier) @name.definition.enum) @definition.enum

; Decorator definitions with decorated class
(export_statement
  decorator: (decorator
    (call_expression
      function: (identifier) @name.definition.decorator))
  declaration: (class_declaration
    name: (type_identifier) @name.definition.decorated_class)) @definition.decorated_class

; Explicitly capture class name in decorated class
(class_declaration
  name: (type_identifier) @name.definition.class) @definition.class

; Namespace declarations
(internal_module
  name: (identifier) @name.definition.namespace) @definition.namespace

; Interface declarations with generic type parameters and constraints
(interface_declaration
  name: (type_identifier) @name.definition.interface
  type_parameters: (type_parameters)?) @definition.interface

; Type alias declarations with generic type parameters and constraints
(type_alias_declaration
  name: (type_identifier) @name.definition.type
  type_parameters: (type_parameters)?) @definition.type

; Utility Types
(type_alias_declaration
  name: (type_identifier) @name.definition.utility_type) @definition.utility_type

; Class Members and Properties
(public_field_definition
  name: (property_identifier) @name.definition.property) @definition.property

; Constructor
(method_definition
  name: (property_identifier) @name.definition.constructor
  (#eq? @name.definition.constructor "constructor")) @definition.constructor

; Getter/Setter Methods
(method_definition
  name: (property_identifier) @name.definition.accessor) @definition.accessor

; Async Functions
(function_declaration
  name: (identifier) @name.definition.async_function) @definition.async_function

; Async Arrow Functions
(variable_declaration
  (variable_declarator
    name: (identifier) @name.definition.async_arrow
    value: (arrow_function))) @definition.async_arrow
`

	tsxQuery = typescriptQuery + `

; React Component Definitions
; Function Components
(function_declaration
  name: (identifier) @name.definition.component) @definition.component

; Arrow Function Components
(variable_declaration
  (variable_declarator
    name: (identifier) @name.definition.component
    value: [(arrow_function) (function_expression)])) @definition.component

; Class Components
(class_declaration
  name: (type_identifier) @name.definition.component
  (class_heritage
    (extends_clause
      (member_expression) @base))) @definition.component

; Higher Order Components
(variable_declaration
  (variable_declarator
    name: (identifier) @name.definition.component
    value: (call_expression
      function: (identifier) @hoc))) @definition.component
  (#match? @hoc "^with[A-Z]")

; Capture all named definitions (component or not)
(variable_declaration
  (variable_declarator
    name: (identifier) @name.definition
    value: [
      (call_expression) @value
      (arrow_function) @value
    ])) @definition.component

; Capture all exported component declarations, including React component wrappers
(export_statement
  (variable_declaration
    (variable_declarator
      name: (identifier) @name.definition.component
      value: [
        (call_expression) @value
        (arrow_function) @value
      ]))) @definition.component

; Capture React component name inside wrapped components
(call_expression
  function: (_)
  arguments: (arguments
    (arrow_function))) @definition.wrapped_component

; HOC definitions - capture both the HOC name and definition
(variable_declaration
  (variable_declarator
    name: (identifier) @name.definition.hoc
    value: (arrow_function
      parameters: (formal_parameters)))) @definition.hoc

; Type definitions (to include interfaces and types)
(type_alias_declaration
  name: (type_identifier) @name.definition.type) @definition.type

(interface_declaration
  name: (type_identifier) @name.definition.interface) @definition.interface

; Enhanced Components
(variable_declaration
  (variable_declarator
    name: (identifier) @name.definition.component
    value: (call_expression))) @definition.component

; Types and Interfaces
(interface_declaration
  name: (type_identifier) @name.definition.interface) @definition.interface

(type_alias_declaration
  name: (type_identifier) @name.definition.type) @definition.type

; JSX Component Usage - Capture all components in JSX
(jsx_element
  open_tag: (jsx_opening_element
    name: [(identifier) @component (member_expression) @component])) @definition.component
  (#match? @component "^[A-Z]")

(jsx_self_closing_element
  name: [(identifier) @component (member_expression) @component]) @definition.component
  (#match? @component "^[A-Z]")

; Capture all identifiers in JSX expressions that start with capital letters
(jsx_expression
  (identifier) @jsx_component) @definition.jsx_component
  (#match? @jsx_component "^[A-Z]")

; Capture all member expressions in JSX
(member_expression
  object: (identifier) @object
  property: (property_identifier) @property) @definition.member_component
  (#match? @object "^[A-Z]")

; Capture components in conditional expressions
(ternary_expression
  consequence: (parenthesized_expression
    (jsx_element
      open_tag: (jsx_opening_element
        name: (identifier) @component)))) @definition.conditional_component
  (#match? @component "^[A-Z]")

(ternary_expression
  alternative: (jsx_self_closing_element
    name: (identifier) @component)) @definition.conditional_component
  (#match? @component "^[A-Z]")

; Enhanced TypeScript Support - React-specific patterns only
; Method Definitions specific to React components
(method_definition
  name: (property_identifier) @name.definition.method) @definition.method

; React Hooks
(variable_declaration
  (variable_declarator
    name: (array_pattern) @name.definition.hook
    value: (call_expression
      function: (identifier) @hook_name))) @definition.hook
  (#match? @hook_name "^use[A-Z]")

; Custom Hooks
(function_declaration
  name: (identifier) @name.definition.custom_hook) @definition.custom_hook
  (#match? @name.definition.custom_hook "^use[A-Z]")

; Context Providers and Consumers
(variable_declaration
  (variable_declarator
    name: (identifier) @name.definition.context
    value: (member_expression))) @definition.context

; React-specific decorators
(decorator) @definition.decorator
`

	pythonQuery = `

; Class definitions
(class_definition
  name: (identifier) @name.definition.class) @definition.class

; Function definitions
(function_definition
  name: (identifier) @name.definition.function) @definition.function

; Method definitions (functions within a class)
(class_definition
  body: (block
    (function_definition
      name: (identifier) @name.definition.method))) @definition.method

; Individual method definitions (to capture all methods)
(class_definition
  body: (block
    (function_definition
      name: (identifier) @name.definition.method_direct))) @definition.method_direct

; Decorated functions and methods
(decorated_definition
  (decorator) @decorator
  definition: (function_definition
    name: (identifier) @name.definition.decorated_function)) @definition.decorated_function

; Decorated classes
(decorated_definition
  (decorator) @decorator
  definition: (class_definition
    name: (identifier) @name.definition.decorated_class)) @definition.decorated_class

; Module-level variables
(expression_statement
  (assignment
    left: (identifier) @name.definition.variable)) @definition.variable

; Constants (uppercase variables by convention)
(expression_statement
  (assignment
    left: (identifier) @name.definition.constant
    (#match? @name.definition.constant "^[A-Z][A-Z0-9_]*$"))) @definition.constant

; Async functions
(function_definition
  "async" @async
  name: (identifier) @name.definition.async_function) @definition.async_function

; Async methods
(class_definition
  body: (block
    (function_definition
      "async" @async
      name: (identifier) @name.definition.async_method))) @definition.async_method

; Lambda functions
(lambda
  parameters: (lambda_parameters) @parameters) @definition.lambda

; Class attributes
(class_definition
  body: (block
    (expression_statement
      (assignment
        left: (identifier) @name.definition.class_attribute)))) @definition.class_attribute

; Property getters/setters (using decorators)
(class_definition
  body: (block
    (decorated_definition
      (decorator
        (call
          function: (identifier) @property
          (#eq? @property "property")))
      definition: (function_definition
        name: (identifier) @name.definition.property_getter)))) @definition.property_getter

; Property setters
(class_definition
  body: (block
    (decorated_definition
      (decorator
        (attribute
          object: (identifier) @property
          attribute: (identifier) @setter
          (#eq? @property "property")
          (#eq? @setter "setter")))
      definition: (function_definition
        name: (identifier) @name.definition.property_setter)))) @definition.property_setter

; Type annotations for variables
(expression_statement
  (assignment
    left: (identifier) @name.definition.typed_variable
    type: (type))) @definition.typed_variable

; Type annotations for function parameters
(typed_parameter
  (identifier) @name.definition.typed_parameter) @definition.typed_parameter

; Direct type annotations for variables (in if __name__ == "__main__" block)
(assignment
  left: (identifier) @name.definition.direct_typed_variable
  type: (type)) @definition.direct_typed_variable

; Type annotations for functions with return type
(function_definition
  name: (identifier) @name.definition.typed_function
  return_type: (type)) @definition.typed_function

; Dataclasses (identified by decorator)
(decorated_definition
  (decorator
    (call
      function: (identifier) @dataclass
      (#eq? @dataclass "dataclass")))
  definition: (class_definition
    name: (identifier) @name.definition.dataclass)) @definition.dataclass

; Nested functions
(function_definition
  body: (block
    (function_definition
      name: (identifier) @name.definition.nested_function))) @definition.nested_function

; Nested classes
(function_definition
  body: (block
    (class_definition
      name: (identifier) @name.definition.nested_class))) @definition.nested_class

; Generator functions (identified by yield)
(function_definition
  name: (identifier) @name.definition.generator_function
  body: (block
    (expression_statement
      (yield)))) @definition.generator_function

; List comprehensions
(expression_statement
  (assignment
    right: (list_comprehension) @name.definition.list_comprehension)) @definition.list_comprehension

; Dictionary comprehensions
(expression_statement
  (assignment
    right: (dictionary_comprehension) @name.definition.dict_comprehension)) @definition.dict_comprehension

; Set comprehensions
(expression_statement
  (assignment
    right: (set_comprehension) @name.definition.set_comprehension)) @definition.set_comprehension

; Direct list comprehensions (in if __name__ == "__main__" block)
(list_comprehension) @definition.direct_list_comprehension

; Direct dictionary comprehensions (in if __name__ == "__main__" block)
(dictionary_comprehension) @definition.direct_dict_comprehension

; Direct set comprehensions (in if __name__ == "__main__" block)
(set_comprehension) @definition.direct_set_comprehension

; Class methods (identified by decorator)
(class_definition
  body: (block
    (decorated_definition
      (decorator
        (call
          function: (identifier) @classmethod
          (#eq? @classmethod "classmethod")))
      definition: (function_definition
        name: (identifier) @name.definition.class_method)))) @definition.class_method

; Static methods (identified by decorator)
(class_definition
  body: (block
    (decorated_definition
      (decorator
        (call
          function: (identifier) @staticmethod
          (#eq? @staticmethod "staticmethod")))
      definition: (function_definition
        name: (identifier) @name.definition.static_method)))) @definition.static_method
`

	goQuery = `
; Function declarations with associated comments
(
  (comment)* @doc
  .
  (function_declaration
    name: (identifier) @name.definition.function) @definition.function
  (#strip! @doc "^//\\s*")
  (#set-adjacent! @doc @definition.function)
)

; Method declarations with associated comments
(
  (comment)* @doc
  .
  (method_declaration
    name: (field_identifier) @name.definition.method) @definition.method
  (#strip! @doc "^//\\s*")
  (#set-adjacent! @doc @definition.method)
)

; Type specifications
(type_spec
  name: (type_identifier) @name.definition.type) @definition.type

; Struct definitions
(type_spec
  name: (type_identifier) @name.definition.struct
  type: (struct_type)) @definition.struct

; Interface definitions
(type_spec
  name: (type_identifier) @name.definition.interface
  type: (interface_type)) @definition.interface

; Constant declarations - single constant
(const_declaration
  (const_spec
    name: (identifier) @name.definition.constant)) @definition.constant

; Constant declarations - multiple constants in a block
(const_spec
  name: (identifier) @name.definition.constant) @definition.constant

; Variable declarations - single variable
(var_declaration
  (var_spec
    name: (identifier) @name.definition.variable)) @definition.variable

; Variable declarations - multiple variables in a block
(var_spec
  name: (identifier) @name.definition.variable) @definition.variable

; Type aliases
(type_spec
  name: (type_identifier) @name.definition.type_alias
  type: (type_identifier)) @definition.type_alias

; Init functions
(function_declaration
  name: (identifier) @name.definition.init_function
  (#eq? @name.definition.init_function "init")) @definition.init_function

; Anonymous functions
(func_literal) @definition.anonymous_function
`

	cppQuery = `
; Struct declarations
(struct_specifier name: (type_identifier) @name.definition.class) @definition.class

; Union declarations
(union_specifier name: (type_identifier) @name.definition.class) @definition.class

; Function declarations
(function_declarator declarator: (identifier) @name.definition.function) @definition.function

; Method declarations (field identifier)
(function_declarator declarator: (field_identifier) @name.definition.function) @definition.function

; Class declarations
(class_specifier name: (type_identifier) @name.definition.class) @definition.class

; Enum declarations
(enum_specifier name: (type_identifier) @name.definition.enum) @definition.enum

; Namespace declarations
(namespace_definition name: (namespace_identifier) @name.definition.namespace) @definition.namespace

; Template declarations
(template_declaration) @definition.template

; Template class declarations
(template_declaration (class_specifier name: (type_identifier) @name.definition.template_class)) @definition.template_class

; Template function declarations
(template_declaration (function_definition declarator: (function_declarator declarator: (identifier) @name.definition.template_function))) @definition.template_function

; Virtual functions
(function_definition (virtual)) @definition.virtual_function

; Auto type deduction
(declaration type: (placeholder_type_specifier (auto))) @definition.auto_variable

; Structured bindings (C++17) - using a text-based match
(declaration) @definition.structured_binding
  (#match? @definition.structured_binding "\\[.*\\]")

; Inline functions and variables - using a text-based match
(function_definition) @definition.inline_function
  (#match? @definition.inline_function "inline")

(declaration) @definition.inline_variable
  (#match? @definition.inline_variable "inline")

; Noexcept specifier - using a text-based match
(function_definition) @definition.noexcept_function
  (#match? @definition.noexcept_function "noexcept")

; Function with default parameters - using a text-based match
(function_declarator) @definition.function_with_default_params
  (#match? @definition.function_with_default_params "=")

; Variadic templates - using a text-based match
(template_declaration) @definition.variadic_template
  (#match? @definition.variadic_template "\\.\\.\\.")

; Explicit template instantiation - using a text-based match
(template_declaration) @definition.template_instantiation
  (#match? @definition.template_instantiation "template\\s+class|template\\s+struct")
`

	cQuery = `
(struct_specifier name: (type_identifier) @name.definition.class body:(_)) @definition.class

(declaration type: (union_specifier name: (type_identifier) @name.definition.class)) @definition.class

(function_declarator declarator: (identifier) @name.definition.function) @definition.function

(type_definition declarator: (type_identifier) @name.definition.type) @definition.type
`

	csharpQuery = `
(class_declaration
 name: (identifier) @name.definition.class
) @definition.class

(interface_declaration
 name: (identifier) @name.definition.interface
) @definition.interface

(method_declaration
 name: (identifier) @name.definition.method
) @definition.method

(namespace_declaration
 name: (identifier) @name.definition.module
) @definition.module
`

	rubyQuery = `
(
  (comment)* @doc
  .
  [
    (method
      name: (_) @name.definition.method) @definition.method
    (singleton_method
      name: (_) @name.definition.method) @definition.method
  ]
  (#strip! @doc "^#\\s*")
  (#select-adjacent! @doc @definition.method)
)

(alias
  name: (_) @name.definition.method) @definition.method

(
  (comment)* @doc
  .
  [
    (class
      name: [
        (constant) @name.definition.class
        (scope_resolution
          name: (_) @name.definition.class)
      ]) @definition.class
    (singleton_class
      value: [
        (constant) @name.definition.class
        (scope_resolution
          name: (_) @name.definition.class)
      ]) @definition.class
  ]
  (#strip! @doc "^#\\s*")
  (#select-adjacent! @doc @definition.class)
)

(
  (module
    name: [
      (constant) @name.definition.module
      (scope_resolution
        name: (_) @name.definition.module)
    ]) @definition.module
)
`

	javaQuery = `
; Class declarations
(class_declaration
  name: (identifier) @name.definition.class) @definition.class

; Method declarations
(method_declaration
  name: (identifier) @name.definition.method) @definition.method

; Interface declarations
(interface_declaration
  name: (identifier) @name.definition.interface) @definition.interface

; Enum declarations
(enum_declaration
  name: (identifier) @name.definition.enum) @definition.enum

; Enum constants
(enum_constant
  name: (identifier) @name.definition.enum_constant) @definition.enum_constant

; Annotation type declarations
(annotation_type_declaration
  name: (identifier) @name.definition.annotation) @definition.annotation

; Field declarations
(field_declaration
  declarator: (variable_declarator
    name: (identifier) @name.definition.field)) @definition.field

; Constructor declarations
(constructor_declaration
  name: (identifier) @name.definition.constructor) @definition.constructor

; Inner class declarations
(class_body
  (class_declaration
    name: (identifier) @name.definition.inner_class)) @definition.inner_class

; Anonymous class declarations
(object_creation_expression
  (class_body)) @definition.anonymous_class

; Lambda expressions
(lambda_expression) @definition.lambda

; Type parameters (for generics)
(type_parameters) @definition.type_parameters

; Package declarations
(package_declaration
  (scoped_identifier) @name.definition.package) @definition.package

; Import declarations
(import_declaration) @definition.import
`

	phpQuery = `
(class_declaration
  name: (name) @name.definition.class) @definition.class

(function_definition
  name: (name) @name.definition.function) @definition.function

(method_declaration
  name: (name) @name.definition.function) @definition.function
`

	swiftQuery = `
(class_declaration
  name: (type_identifier) @name) @definition.class

(protocol_declaration
  name: (type_identifier) @name) @definition.interface

(class_declaration
    (class_body
        [
            (function_declaration
                name: (simple_identifier) @name
            )
            (subscript_declaration
                (parameter (simple_identifier) @name)
            )
            (init_declaration "init" @name)
            (deinit_declaration "deinit" @name)
        ]
    )
) @definition.method

(class_declaration
    (class_body
        [
            (property_declaration
                (pattern (simple_identifier) @name)
            )
        ]
    )
) @definition.property

(property_declaration
    (pattern (simple_identifier) @name)
) @definition.property

(function_declaration
    name: (simple_identifier) @name) @definition.function
`

	kotlinQuery = `
(class_declaration
  (type_identifier) @name.definition.class
) @definition.class

(function_declaration
  (simple_identifier) @name.definition.function
) @definition.function

(object_declaration
  (type_identifier) @name.definition.object
) @definition.object

(property_declaration
  (simple_identifier) @name.definition.property
) @definition.property

(type_alias
  (type_identifier) @name.definition.type
) @definition.type
`
)
