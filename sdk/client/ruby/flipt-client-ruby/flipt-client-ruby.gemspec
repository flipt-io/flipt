require_relative 'lib/client/version'

Gem::Specification.new do |spec|
  spec.name          = "flipt_client"
  spec.version       = Flipt::Client::VERSION
  spec.authors       = ["Flipt Devs"]
  spec.email         = ["dev@flipt.io"]

  spec.summary       = "Ruby Client SDK for Flipt"
  spec.description   = "..."
  spec.homepage      = "https://www.flipt.io"
  spec.license       = "MIT"
  spec.required_ruby_version = Gem::Requirement.new(">= 2.3.0")

  spec.metadata["allowed_push_host"] = "TODO: Set to 'http://mygemserver.com'"

  spec.metadata["homepage_uri"] = spec.homepage
  # spec.metadata["source_code_uri"] = "TODO: Put your gem's public repo URL here."
  # spec.metadata["changelog_uri"] = "TODO: Put your gem's CHANGELOG.md URL here."

  spec.files       = Dir.glob("{lib,spec}/**/*") + ["README.md"]
  spec.bindir        = "exe"
  spec.executables   = spec.files.grep(%r{^exe/}) { |f| File.basename(f) }
  spec.require_paths = ["lib"]
end
