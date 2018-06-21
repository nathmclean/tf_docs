# tf_docs

Parses Terraform modules extracting useful information to use as documentation

## Usage

`modules, err := tf.ParseModules("path/to/my/modules")`

## How it Works

### Description

Looks for a comment on line 1 of a .tf file that starts with the module name (directory name):

```
# moduleName does this...
# also does this
```

```
\* moduleName does a bunch of things
*\
```

### Variables

Takes the variable name, type and description. If the default is present then extracted too.
A required field is calculated based on whether a default is available.

```
variable "name" {
  type        = "type"
  description = "description"
  default     = "default" # optional
}
```

### Outputs

The name and description of outputs are extracted

```
output "name" {
  description = "description"
  value       = "value"
}
```

### Modules and Resources

Leading comments (on the line preceding the declaration of the module/resource) are used as descriptions.

```
// describe this resource
resource "type" "name" {
   ...
}
```

#### Resources

The type and name of the resource are extracted.

```
resource "name" "type {
  ...
}
```

#### Modules

The name and source of the modules are extracted.

```
module "name" {
  source = "source"
}
```