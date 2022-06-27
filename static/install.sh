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

  if grep -q "/.cubby/bin" "$HOME/$rcfile"; then
    echo "Cubby already in zsh path; no need to add path"
  else
    echo "Adding Cubby to PATH in $rcfile";
    echo 'export PATH=$PATH:$HOME/.cubby/bin' >> $HOME/$rcfile;
  fi
}

install_binary() {
  isnew="no"
  if [[ ! -d $HOME/.cubby ]]; then
    echo "Making $HOME/.cubby and writing configs..."
    mkdir $HOME/.cubby;
    write_config;
    mkdir $HOME/.cubby/bin;
    isnew="yes"
  else
    echo "Leaving old Cubby configs in tact...";
  fi
  echo "Downloading Cubby binary at: $1";
  curl -o $HOME/.cubby/bin/cubby -s -S -L "$1"
  if [ $? -ne 0 ]; then
    echo "Failed to download binary.";
    exit 1;
  fi
  chmod +x $HOME/.cubby/bin/cubby;
  add_to_path;
  if [[ "$isnew" = "yes" ]]; then
    read -p "Do you need to register a new user account? (y/n)  " newacct
    if [[ newacct == "y"* ]]; then
      echo "Registering new account...";
      if ! $HOME/.cubby/bin/cubby signup; then
        printf "\nDO YOU ALREADY HAVE AN ACCOUNT?";
        printf "We couldn't register a new account with that email address. Usually that's because one already exists. Please check your ~/.cubby/cubby-client.yaml file for accuracy and run \"cubby signup\" to try again.\n";
        echo "You will get this error if you have already signed up for a Cubby account with this email address. If that's the case, you can ignore this warning and begin using Cubby (you'll need to restart your shell for PATH changes to take effect).";
        exit 1;
      else
        printf "\nSign up was successful! Please restart your shell for PATH change to take effect. After that, you're\n";
        echo "ready to start using Cubby! Please see our README on Github for ideas of where to start. 😁";
      fi
    else
      printf "\nOK, your config has been set with the credentials you provided, but we didn't registered them as a new account. You can start using Cubby now!\n";
      echo "If this was a mistake and you indeed intended to register these credentials as a new account, fear not: a quick \"cubby signup\" will fix that.";
    fi
  else
    echo "Your binary has been updated and all configs were left in tact. Happy Cubbyholing!. 😁 ";
  fi
}

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
       MACHINE_TYPE=`uname -m`
       if [ ${MACHINE_TYPE} == 'x86_64' ]; then
          # 64-bit stuff here
          echo "Running Linux x64 installer...";
          install_binary "https://www.cubbycli.com/static/dist/cubby_linux_amd64"
       else
          echo "Only x64 is supported with the auto-install script. Please build from source until an installer is available.";
          exit 1;
       fi
elif [[ "$OSTYPE" == "darwin"* ]]; then
        # Mac OSX
       MACHINE_TYPE=`uname -m`
       if [ ${MACHINE_TYPE} == 'x86_64' ]; then
          # 64-bit stuff here
          echo "Running Mac OS X x64 installer...";
          install_binary "https://www.cubbycli.com/static/dist/cubby_darwin_amd64"
       else
          # echo "Only x64 is supported with the auto-install script. Please build from source until an installer is available.";
          echo "Running Mac OS X ARM64 installer...";
          install_binary "https://www.cubbycli.com/static/dist/cubby_darwin_arm64"
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

