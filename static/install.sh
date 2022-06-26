#!/bin/bash

write_config() {
  read -p "Enter password: " password 
  read -p "Enter an encryption passphrase: " encpass 
  read -p "Enter an email address to associate with your account: " email
  echo "Using email ${email}, password ${password}, encryption passphrase ${encpass}";
  cat  > $HOME/cubby-client.yaml << EndOfMessage
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

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
       echo "Running Linux installer...";
elif [[ "$OSTYPE" == "darwin"* ]]; then
        # Mac OSX
       MACHINE_TYPE=`uname -m`
       if [ ${MACHINE_TYPE} == 'x86_64' ]; then
          # 64-bit stuff here
          echo "Running MacOS X x64 installer...";
          write_config;
       else
          # 32-bit stuff here
          echo "Only x64 is supported currently. Please build from source until an installer is available.";
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

