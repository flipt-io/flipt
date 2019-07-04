workflow "On Push" {
  resolves = [
    "Publish Docs",
    "On Tag",
    "clarkio/snyk-cli-action@master",
  ]
  on = "push"
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

action "clarkio/snyk-cli-action@master" {
  uses = "clarkio/snyk-cli-action@master"
  secrets = ["SNYK_TOKEN"]
}
