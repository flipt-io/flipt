workflow "Publish Docs on Release" {
  on = "push"
  resolves = [
    "Publish Docs",
    "On Tag",
  ]
}

action "On Tag" {
  uses = "actions/bin/filter@3c0b4f0e63ea54ea5df2914b4fabf383368cd0da"
  args = "tag"
}

action "Publish Docs" {
  uses = "./actions/publish-docs"
  needs = ["On Tag"]
  secrets = ["GITHUB_TOKEN"]
}
