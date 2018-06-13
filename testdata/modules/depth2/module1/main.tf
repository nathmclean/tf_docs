// module1 is a test module

variable "test" {
  type        = "string"
  description = "this is a variable"
}

output "test" {
  value       = "val"
  description = "output description"
}

// this is a resource
resource "aws_ami" "ami" {
  name = "ami-ascdefg"
}

// here's a module
module "test" {
  source = "../"
}
