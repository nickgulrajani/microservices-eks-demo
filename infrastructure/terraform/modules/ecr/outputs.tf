output "repository_urls" {
  description = "Map of repository names to URLs"
  value = {
    for k, v in aws_ecr_repository.repositories : k => v.repository_url
  }
}

output "repository_arns" {
  description = "Map of repository names to ARNs"
  value = {
    for k, v in aws_ecr_repository.repositories : k => v.arn
  }
}
