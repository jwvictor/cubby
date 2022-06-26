#!/bin/bash

set -e

write_config() {
  read -p "Enter password: " password 
  read -p "Enter an encryption passphrase: " encpass 
  read -p "Enter an email address to associate with your account: " email
  echo "Using email ${email}, password ${password}, encryption passphrase ${encpass}";
  cat  > $HOME/.cubby/cubby-client.yaml << EndOfMessage
host: https://public.cubbycli.com
port: 443
options:
  viewer: editor
  body-only: true
user:
  email: ${email} 
  password: ${password}
crypto:
  symmetric-key: ${encpass}
  mode: symmetric
EndOfMessage
}

add_to_path() {
  shell="$SHELL";
  rcfile=".bashrc"
  if [[ "$shell" == *"zsh" ]]; then
    rcfile=".zshrc";
  fi

  echo "shell is $shell, rcfile is $rcfile"

  if grep -q "/.cubby/bin" "$HOME/$rcfile"; then
    echo "Cubby already in zsh path; no need to add path"
  else
    echo "Adding cubby to zsh path in $rcfile";
    echo 'export PATH=$PATH:$HOME/.cubby/bin' >> $HOME/$rcfile;
  fi
}

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
       echo "Running Linux installer...";
elif [[ "$OSTYPE" == "darwin"* ]]; then
        # Mac OSX
       MACHINE_TYPE=`uname -m`
       if [ ${MACHINE_TYPE} == 'x86_64' ]; then
          # 64-bit stuff here
          echo "Running MacOS X x64 installer...";
          mkdir $HOME/.cubby;
          write_config;
          mkdir $HOME/.cubby/bin;
          curl -o $HOME/.cubby/bin/cubby "https://www.cubbycli.com/static/dist/cubby_darwin_amd64"
          if [ $? -ne 0 ]; then
            echo "Failed to download binary.";
            exit 1;
          fi
          add_to_path;
          chmod +x $HOME/.cubby/bin/cubby;
          echo "Wrote configuration file, running signup with \"cubby signup\"..."
          $HOME/.cubby/bin/cubby signup;
          if [ $? -ne 0 ]; then
            echo "Sign up failed. Please check your ~/.cubby/cubby-client.yaml file for accuracy and run \"cubby signup\" to try again.";
          else
            echo "Sign up was successful! Please restart your shell for PATH change to take effect. After that, you're";
            echo "ready to start using Cubby! Please see our README on Github for ideas of where to start. 😁";
          fi
       else
          # 32-bit stuff here
          echo "Only x64 is supported with the auto-install script. Please build from source until an installer is available.";
         exit 1;
       fi
elif [[ "$OSTYPE" == "cygwin" ]]; then
        # POSIX compatibility layer and Linux environment emulation for Windows
       echo "Cygwin not yet supported.";
       exit 1;
elif [[ "$OSTYPE" == "msys" ]]; then
        # Lightweight shell and GNU utilities compiled for Windows (part of MinGW)
       echo "Msys not yet supported.";
       exit 1;
elif [[ "$OSTYPE" == "win32" ]]; then
        # I'm not sure this can happen.
       echo "Win32 not yet supported.";
       exit 1;
elif [[ "$OSTYPE" == "freebsd"* ]]; then
       echo "FreeBSD not yet supported.";
       exit 1;
        # ...
else
        # Unknown.
       echo "OS not yet supported.";
       exit 1;
fi

