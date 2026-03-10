{ pkgs, ... }:

{
  # https://devenv.sh/packages/
  packages = [ pkgs.git pkgs.mise pkgs.gcc pkgs.nodejs ];

  # https://devenv.sh/scripts/
  scripts.hello.exec = "echo 'hello from Flipt v2!'";

  # https://devenv.sh/languages/
  languages.go = {
    enable = true;
    package = pkgs.go_1_26;
  };
  languages.typescript.enable = true;

  # https://devenv.sh/pre-commit-hooks/
  # pre-commit.hooks.shellcheck.enable = true;

  # https://devenv.sh/processes/
  processes = {
    backend.exec = "mise run dev";
    backend.notify.enable = true;
    frontend.exec = "mise run ui:dev";
  };

  # See full reference at https://devenv.sh/reference/options/
}
