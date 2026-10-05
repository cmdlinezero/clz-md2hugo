with import <nixpkgs> {};

let
  scriptZellijLayout =
    pkgs.writeText "dev.kdl" (builtins.readFile ./dev.kdl);

  zellijConfig = pkgs.writeText "config.kdl" ''
    default_mode "normal"
  '';

  zellijLayout = pkgs.writeShellScriptBin "dev-zellij" ''
    ${pkgs.zellij}/bin/zellij \
      --config ${zellijConfig} \
      --layout ${scriptZellijLayout}
  '';

in pkgs.mkShell {
  name = "go-dev";

  nativeBuildInputs = with pkgs; [
    go
    tree
    zsh
    zellij
    zellijLayout
    oh-my-zsh
  ];

  LANGUAGE = "Go";
  VERSION = "go version";

  shellHook = ''
    echo "Welcome to $LANGUAGE Development Environment"
    $VERSION

    export ZSH=${pkgs.oh-my-zsh}/share/oh-my-zsh

    exec dev-zellij
  '';
}
