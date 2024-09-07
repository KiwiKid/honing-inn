{
  description = "A flake for building, running, and deploying a Go program with devShells";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";
    nixpkgs-unstable.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs, nixpkgs-unstable }: 
    let 
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
      unstablePkgs = import nixpkgs-unstable { inherit system; };
      go = unstablePkgs.go_1_21;
      templ = unstablePkgs.templ;
      air = unstablePkgs.air;
      docker = unstablePkgs.docker;
      flyctl = unstablePkgs.flyctl;
    in
    {
      devShells = {
        build = pkgs.mkShell {
          buildInputs = [
            go
            templ
            flyctl
          ];
          shellHook = ''
            echo "Building the Go project..."
            cd ~/mine/honing-inn
            git config user.name $GIT_AUTHOR_USER
            git config user.email $GIT_AUTHOR_EMAIL
            export DATABASE_URL=./tmp/dev.db
            export IMAGE_DIR=./tmp/images
            templ generate && go build -o ./tmp/main .
          '';
        };

        deploy = pkgs.mkShell {
          buildInputs = [
            go
            templ
            flyctl
          ];
          shellHook = ''
            echo "Building the Go project..."
            cd ~/mine/honing-inn
            fly deploy
          '';
        };

        tdeploy = pkgs.mkShell {
          buildInputs = [
            go
            templ
            flyctl
          ];
          shellHook = ''
            echo "Building the Go project..."
            git config user.name $GIT_AUTHOR_USER
            git config user.email $GIT_AUTHOR_EMAIL
            export DATABASE_URL=./tmp/data/dev.db
            templ generate
            echo "Deploying the application..."
            fly deploy -c fly-dev.toml
          '';
        };

        dev = pkgs.mkShell {
          buildInputs = [ 
            air
            templ
            go
            pkgs.tmux
            flyctl
            pkgs.golint
          ];
          shellHook = ''
            git config user.name $GIT_AUTHOR_USER
            git config user.email $GIT_AUTHOR_EMAIL
           export  DATABASE_URL=./dev.db
           export IMAGE_DIR=./tmp/images

            export PASS=pass
            code .
            tmux kill-session -t devSession
            tmux new-session -d -s devSession \; \
              split-window -h \; \
              send-keys -t 0 'templ generate --watch' C-m \; \
              send-keys -t 1 'air' C-m \; \
            attach-session -t devSession
          '';
          shellExit = ''
            tmux kill-session -t devSession
          '';
        };

        shell = pkgs.mkShell {
          buildInputs = [ 
            air
            templ
            go
            pkgs.tmux
            flyctl
            pkgs.golint
          ];
          shellHook = ''
            cd ~/mine/honing-inn
          '';
        };

        dockerBuild = pkgs.mkShell {
          buildInputs = [ 
            docker
            flyctl
          
          ];
          shellHook = ''
            echo "Building Docker image..."
            docker build -t baileys-hammer .
            docker run -p 8080:8080 baileys-hammer
          '';
        };


      };
      defaultPackage.x86_64-linux = self.devShells.dev;

    };
}
