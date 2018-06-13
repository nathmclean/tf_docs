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
