{ pkgs, ... }:

{
  # https://devenv.sh/basics/
  # env.GREET = "devenv";

  # https://devenv.sh/packages/
  packages = [ pkgs.git pkgs.mage pkgs.gcc pkgs.sqlite pkgs.nodejs ];

  # https://devenv.sh/scripts/
  scripts.hello.exec = "echo 'hello from Flipt!'";

  # https://devenv.sh/languages/
  languages.go.enable = true;
  languages.typescript.enable = true;

  # https://devenv.sh/pre-commit-hooks/
  # pre-commit.hooks.shellcheck.enable = true;

  # https://devenv.sh/processes/
  processes = {
    backend.exec = "mage dev";
    frontend.exec = "mage ui:dev";
  };

  # See full reference at https://devenv.sh/reference/options/
}
