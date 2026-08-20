{ config, lib, pkgs, ... }:

# NixOS module for linny-mcp.
#
# SCAFFOLD STATE: the option interface below is stable and complete, but the
# hardened systemd unit is intentionally minimal here. Milestone 06
# (nixos-module-hardened) adds the full hardening (ProtectSystem=strict,
# RestrictAddressFamilies, SystemCallFilter=@system-service, CapabilityBoundingSet=,
# MemoryDenyWriteExecute, etc.).
#
# SECRET HYGIENE: `tokensFile` is a PATH, never a token value. Never place a
# token literal in a Nix option — options land world-readable in /nix/store.
# Point tokensFile at e.g. config.age.secrets.linny-mcp-tokens.path.

let
  cfg = config.services.linny-mcp;
in
{
  options.services.linny-mcp = {
    enable = lib.mkEnableOption "the linny-mcp MCP server";

    package = lib.mkOption {
      type = lib.types.package;
      default = pkgs.linny-mcp;
      defaultText = lib.literalExpression "pkgs.linny-mcp";
      description = "The linny-mcp package to run (add overlays.default to get it).";
    };

    corpusPath = lib.mkOption {
      type = lib.types.path;
      description = "Path to the Linny notebook (markdown corpus) git working tree.";
    };

    stateDir = lib.mkOption {
      type = lib.types.str;
      default = "/var/lib/linny-mcp";
      description = "Directory for the disposable SQLite/FTS5 index. Safe to delete.";
    };

    listenAddress = lib.mkOption {
      type = lib.types.str;
      default = "127.0.0.1";
      description = ''
        Address to bind. The server refuses to start on a non-loopback,
        non-mesh address without an explicit override. TLS is terminated
        upstream (Hetzner reverse proxy).
      '';
    };

    port = lib.mkOption {
      type = lib.types.port;
      default = 8765;
      description = "TCP port to listen on.";
    };

    tokensFile = lib.mkOption {
      type = lib.types.path;
      description = ''
        Path to the bearer-token file (NOT a token value). Source it from a
        secret manager, e.g. config.age.secrets.linny-mcp-tokens.path, with
        owner set to the service user.
      '';
    };

    user = lib.mkOption {
      type = lib.types.str;
      default = "linny-mcp";
      description = "User to run the service as.";
    };

    group = lib.mkOption {
      type = lib.types.str;
      default = "linny-mcp";
      description = "Group to run the service as.";
    };

    logLevel = lib.mkOption {
      type = lib.types.enum [ "debug" "info" "warn" "error" ];
      default = "info";
      description = "Log verbosity.";
    };

    readOnly = lib.mkOption {
      type = lib.types.bool;
      default = false;
      description = "Force read-only mode regardless of git working-tree state.";
    };
  };

  config = lib.mkIf cfg.enable {
    # Minimal service. Hardening is layered on in milestone 06.
    users.users.${cfg.user} = lib.mkIf (cfg.user == "linny-mcp") {
      isSystemUser = true;
      group = cfg.group;
      home = cfg.stateDir;
    };
    users.groups.${cfg.group} = lib.mkIf (cfg.group == "linny-mcp") { };

    systemd.services.linny-mcp = {
      description = "linny-mcp MCP server";
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];
      serviceConfig = {
        ExecStart = lib.escapeShellArgs ([
          (lib.getExe cfg.package)
          "serve"
          "--corpus" cfg.corpusPath
          "--state-dir" cfg.stateDir
          "--listen" cfg.listenAddress
          "--port" (toString cfg.port)
          "--tokens-file" cfg.tokensFile
          "--log-level" cfg.logLevel
        ] ++ lib.optionals cfg.readOnly [ "--read-only" ]);
        User = cfg.user;
        Group = cfg.group;
        StateDirectory = "linny-mcp";
        Restart = "on-failure";
      };
    };
  };
}
