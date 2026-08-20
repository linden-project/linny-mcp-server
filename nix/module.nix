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
      type = lib.types.nullOr lib.types.path;
      default = null;
      description = ''
        Single-notebook convenience: path to the Linny notebook (markdown corpus)
        git working tree. Desugars to one notebook named "default". Leave null and
        use `notebooks` for multi-notebook setups.
      '';
    };

    stateDir = lib.mkOption {
      type = lib.types.str;
      default = "/var/lib/linny-mcp";
      description = "State dir for the single-notebook convenience path. Disposable index; safe to delete.";
    };

    notebooks = lib.mkOption {
      default = [ ];
      description = ''
        The notebooks to serve. Designed for N notebooks (e.g. personal vs.
        business). If empty, `corpusPath`/`stateDir` desugar to a single
        "default" notebook.
      '';
      type = lib.types.listOf (lib.types.submodule {
        options = {
          name = lib.mkOption { type = lib.types.str; description = "Unique notebook name."; };
          corpusPath = lib.mkOption { type = lib.types.path; description = "Corpus git working tree."; };
          stateDir = lib.mkOption {
            type = lib.types.str;
            description = "Disposable index/state dir for this notebook.";
          };
        };
      });
    };

    publicHostname = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "secondbrain.example.com";
      description = ''
        The public hostname the reverse proxy fronts this server with. Purely
        configuration — no hostname is compiled into the binary. Leave null if unused.
      '';
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

    ntfyTopicURL = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "https://ntfy.example.com/linny-mcp";
      description = ''
        Self-hosted ntfy topic URL for out-of-band degraded-mode alerts. A plain
        POST target (not a secret), so it is fine in a Nix option. Leave null to
        disable alerting.
      '';
    };
  };

  config = lib.mkIf cfg.enable (
    let
      # Desugar the single-notebook convenience into the notebook list.
      resolvedNotebooks =
        if cfg.notebooks != [ ]
        then cfg.notebooks
        else [{ name = "default"; corpusPath = cfg.corpusPath; stateDir = cfg.stateDir; }];

      # Generated config JSON. Contains only PATHS and non-secret settings —
      # never a token value — so it is safe in the world-readable /nix/store.
      configJSON = {
        listenAddress = cfg.listenAddress;
        port = cfg.port;
        tokensFile = toString cfg.tokensFile;
        logLevel = cfg.logLevel;
        readOnly = cfg.readOnly;
        publicHostname = if cfg.publicHostname == null then "" else cfg.publicHostname;
        ntfyTopicURL = if cfg.ntfyTopicURL == null then "" else cfg.ntfyTopicURL;
        notebooks = map (nb: {
          name = nb.name;
          corpusPath = toString nb.corpusPath;
          stateDir = nb.stateDir;
        }) resolvedNotebooks;
      };
      configFile = pkgs.writeText "linny-mcp-config.json" (builtins.toJSON configJSON);

      rwPaths = map (nb: nb.corpusPath) resolvedNotebooks
        ++ map (nb: nb.stateDir) resolvedNotebooks;
    in
    {
      assertions = [{
        assertion = cfg.notebooks != [ ] || cfg.corpusPath != null;
        message = "services.linny-mcp: set either `corpusPath` (single notebook) or `notebooks` (one or more).";
      }];

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
          ExecStart = lib.escapeShellArgs [ (lib.getExe cfg.package) "serve" "--config" configFile ];
          User = cfg.user;
          Group = cfg.group;
          StateDirectory = "linny-mcp";
          Restart = "on-failure";

          # Hardening (briefing §9). The server only needs to read the corpus and
          # write its state dir + the corpus; everything else is sealed off.
          ProtectSystem = "strict";
          ReadWritePaths = rwPaths;
          ProtectHome = true;
          PrivateTmp = true;
          PrivateDevices = true;
          NoNewPrivileges = true;
          RestrictAddressFamilies = [ "AF_INET" "AF_INET6" "AF_UNIX" ];
          SystemCallFilter = [ "@system-service" "~@privileged" ];
          SystemCallArchitectures = "native";
          CapabilityBoundingSet = [ "" ];
          AmbientCapabilities = [ "" ];
          LockPersonality = true;
          MemoryDenyWriteExecute = true;
          ProtectKernelTunables = true;
          ProtectKernelModules = true;
          ProtectKernelLogs = true;
          ProtectControlGroups = true;
          ProtectClock = true;
          ProtectProc = "invisible";
          ProcSubset = "pid";
          RestrictNamespaces = true;
          RestrictSUIDSGID = true;
          RestrictRealtime = true;
          UMask = "0077";
        };
      };
    }
  );
}
