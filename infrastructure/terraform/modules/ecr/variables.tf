variable "project_name" {
  description = "Name of the project"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "repositories" {
  description = "List of ECR repository names to create"
  type        = list(string)
}
