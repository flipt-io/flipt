workflow "Publish Docs On Tag" {
  resolves = [
    "Publish Docs",
    "On Tag",
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

workflow "Run Snyk On Push" {
  resolves = ["clarkio/snyk-cli-action@master"]
  on = "push"
}

action "clarkio/snyk-cli-action@master" {
  uses = "clarkio/snyk-cli-action@master"
  secrets = ["SNYK_TOKEN"]
  args = "test"
}
