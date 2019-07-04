workflow "Publish Docs" {
  on = "push"
  resolves = ["./actions/publish-docs"]
}

action "./actions/publish-docs" {
  uses = "./actions/publish-docs"
  secrets = ["GITHUB_TOKEN"]
}
