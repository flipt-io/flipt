{ pkgs, ... }:

{
  # https://devenv.sh/packages/
  packages = [ pkgs.git pkgs.mage pkgs.gcc pkgs.nodejs ];

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
    backend.exec = "mage -keep dev";
    backend.notify.enable = true;
    frontend.exec = "mage -keep ui:dev";
  };

  # https://devenv.sh/tasks/
  tasks."app:cleanup" = {
    exec = "rm mage_output_file.go";
    after = [ "devenv:processes:backend" ];
  };

  # See full reference at https://devenv.sh/reference/options/
}
