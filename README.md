# golem

golem is a lightweight Minecraft server proxy with autostart/stop written in Go.

## Usage

    Usage of golem:
      -debug
            Log all traffic
      -playersMax int
            Maximum number of players (to display in status message) (default 20)
      -proxyAddr string
            Proxy server address (default ":25565")
      -serverAddr string
            Minecraft server address (default ":25566")
      -serverDirectory string
            Minecraft server working directory
      -serverStart string
            Minecraft start command. Empty disables autostart/stop
      -stopTimeout int
            Wait period to stop server after last disconnect (seconds) (default 60)
      -versionName string
            Minecraft version name (default "1.17.1")
      -versionProtocol int
            Minecraft protocol version (default 756)

## Appendix

### Codebase

- `protocol` provides a wrapper of `net.Conn` that implements the Minecraft
  protocol.
- `proxy` provides a `Proxy` which intercepts and forwards packets in the
  Minecraft protocol and orchestrates server management.
- `server` defines an interface `Server` for a server manager (start, stop,
  execute commands) and implements a basic manager which does no managing.
    - `server/process` implements a server manager by supervising a child
      process.
    - Future server managers can be implemented such as for a remote process or
      a Docker container.

### Distribution on NixOS

The following configuration declaratively enables/disables running Minecraft
behind a proxy with the `proxy.enable` flag. In this example, `./golem` is a
local package built by `buildGoModule`, where the source points to either this
repository via `fetchFromGithub` or a tarball made by `make archive`.

```nix
let
  server-port = 25565; # port to expose (external) minecraft server

  proxy = {
    enable = true;
    package = pkgs.callPackage ./golem { };
    server-port = 25566; # port to run (internal) minecraft server
  };
in
{
  services.minecraft-server = {
    enable = true;
    eula = true;
    declarative = true;
    serverProperties = {
      server-port = (if proxy.enable then proxy.server-port else server-port);
      # ...
    };
  };

  systemd.services.minecraft-server = {
    serviceConfig.ExecStart = lib.mkIf proxy.enable (lib.mkForce
      ''${proxy.package}/bin/golem \
        -proxyAddr :${toString server-port} \
        -serverAddr :${toString proxy.server-port} \
        -serverDirectory ${config.services.minecraft-server.dataDir} \
        -serverStart "${config.services.minecraft-server.package}/bin/minecraft-server ${config.services.minecraft-server.jvmOpts}"
      '');
  };
}
```
