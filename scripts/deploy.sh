#!/bin/bash

echo "Staring project deploy"

# This is straightforward deploy to linux server that does the job for me for now.
# Probably later I'll need to add a hook to deploy on merge to main branch

# Note to windows users: I used to use this script on linux, but it works surprisingly as expected on windows
#   Only what you need is Git Bash terminal which comes with Git for Windows installation
#     If you don't have Git Bash yet, you can get from Git for Windows: https://git-scm.com/download/win


# how to use:
# 1. copy this file to ../bin or elsewhere you like (../bin is convenient because it is in project dir and out of git)
# 2. set variables that is just right below
#     note that some config values is out of variables,
#     jump to "Writing config" line and make sure that everything is correct there
# 3. run this script

# remote's requirements:
# 1. supervisor must be installed and config for supervisor to run server must be created
#     (look at ../scripts/supervisor.conf for example)
#    if you use something else, delete lines here that stop and start supervisor
# 2. postgres database configured and running

# Note: during deploy you'll need to authenticate at remote twice
#   (1. uploading files with scp; 2. setting up with ssh)
#   if you do deploy often, it is obviously useful to do authentication automatically using authentication keys
#   to do so follow steps from http://www.linuxproblem.org/art_9.html
#   That guide will help you generate a pair of authentication keys
#   so you can deploy seamlessly

# set variables required for deploy

# server's host & port
serverHost=refto.dev
serverPort=443
# local project's dir to deploy from
projectDir="/path/to/source"
# remote project's dir to deploy to
remoteProjectDir="/path/to/dir/at/server"
# remote's user
remoteUser=
# remote's addr
remoreAddr=
# database conf
dbAddr=localhost:5432
dbName=refto
dbUser=postgres
dbPassword=postgres

# Github's app to connect users with
# https://github.com/settings/applications/new
githubClientID=
githubClientSecret=

linterPath="golangci-lint"
# end of variables setup

# necessary checks
if [ -z "$remoreAddr"  ]
then
  echo " - Unable to deploy: Remote addr is not set"
  exit
fi
if [ -z "$remoteUser"  ]
then
  echo " - Unable to deploy: Remote user is not set"
  exit
fi
if [ -z "$dbAddr"  ]
then
  echo " - Unable to deploy: Database addr is not set"
  exit
fi
if [ -z "$dbName"  ]
then
  echo " - Unable to deploy: Database name is not set"
  exit
fi
if [ -z "$dbUser"  ]
then
  echo " - Unable to deploy: Database user is not set"
  exit
fi
if [ -z "$dbPassword"  ]
then
  echo " - Unable to deploy: Database password is not set"
  exit
fi
if [ -z "$serverHost"  ]
then
  echo " - Unable to deploy: Server's host is not set"
  exit
fi
if [ -z "$githubClientID"  ]
then
  echo " - Github client ID is not set!"
  exit
fi
if [ -z "$githubClientSecret"  ]
then
  echo " - Github client secret is not set!"
  exit
fi
if [ -z "$linterPath" ]
then
  echo " - linter's path is not set!"
  echo " - For installation refer to : https://golangci-lint.run/usage/install/#linux-and-windows"
  exit
fi

cd "$projectDir" || (echo "Unable to change dir to $projectDir (check projectDir variable)"; exit)
echo " - Working DIR is set to: $projectDir"

# check for linter installation
if ! command -v "$linterPath" &> /dev/null
then
  echo " - linter is not found at '$linterPath'"
  echo " - For installation refer to:"
  echo "   https://golangci-lint.run/usage/install/#linux-and-windows"
  echo " - If you already installed golangci-lint, make sure that \$linterPath variable is correct"
  exit
fi

# linter's error message with a bit of randomness just for fun
# not sure if messages is appropriate for you (or correct), please fix it by yourself if something went wrong
errText[0]="Fix above errors, young apprentice"
errText[1]="Whoops, something went wrong"
errText[2]="Whoops, you are welcomed by company of noobs ^^"
errText[3]="Whoops, get out from error loops!"
errText[4]="Whoops, your code poops..."
errText[5]="Whoops, you'd better not be distracted by boobs"
errText[6]="Unable to deploy: Please fix above error(s)"
errText[7]="Argh! Sum sing vent rong!"
errText[8]="Deal with errors || run. That's it."
errText[9]="fix_errors() || die();"
errText[10]="Не who makes no mistakes, makes nothing"
errText[11]="Do something else for a while"
errText[12]="If you can't deal with this error(s), don't hesitate to ask at {linkToWhereWeCanTalk}"
# {linkToWhereWeCanTalk} is not a joke, I just don't have a place or platform for that (for now)
errText[13]="WTF?!"
errText[14]="Not again..."

echo " - Linting hard with care..."
$linterPath run || { echo " "; echo " - ${errText[(($RANDOM % ${#errText[@]}))]}"; exit; }

# test api
cd ./server/test || exit
echo " - Testing API..."
go test || { echo " - Unable to deploy: tests failed"; exit; }

echo " - Compiling API server..."
cd "$projectDir" || exit
env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o ./bin/refto-server cmd/server/main.go || exit

echo " - Compiling CLI..."
env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o ./bin/refto-cli cmd/cli/main.go || exit

echo " - Making archive of static data..."
tar -czf refto-static.tar.gz ./web

echo " - Copying files to remote (${remoteUser}@${remoreAddr})..."
scp ./bin/refto-server ./bin/refto-cli refto-static.tar.gz "${remoteUser}@${remoreAddr}":~/ || exit

echo " - Setting up server on remote..."
ssh -T "${remoteUser}@${remoreAddr}" << EOF
echo " - Stopping supervisor..."
sudo service supervisor stop || exit

echo " - Extracting static data..."
tar -xzf ~/refto-static.tar.gz -C $remoteProjectDir || exit
rm -f ~/refto-static.tar.gz || exit

echo " - Moving binaries..."
mv ~/refto-server $remoteProjectDir/server || exit
mv ~/refto-cli $remoteProjectDir/cli || exit
chmod +x $remoteProjectDir/server
chmod +x $remoteProjectDir/cli

echo " - Writing config..."
/bin/cat <<EOM > $remoteProjectDir/.config.yaml
app_env: release

db:
  addr: $dbAddr
  user: $dbUser
  password: $dbPassword
  name: $dbName
  log_queries: true

server:
  host: $serverHost
  port: $serverPort
  api_base_path: api
  static:
    local: "./web"
    web: "/~/"

github:
  client_id: $githubClientID
  client_secret: $githubClientSecret
  data_warden:
    app_id: 1
    install_id: 1
    pem_path: "private-key.pem"

dir:
  data: "$remoteProjectDir/data/"
  logs: ""
EOM

echo " - Migrating database..."
cd $remoteProjectDir || exit
./cli migrate || { echo " - Failed to migrate"; exit; }

echo " - Starting supervisor..."
sudo service supervisor start || exit
EOF

# cleanup
rm -f refto-static.tar.gz || exit
rm -f ./bin/refto-server || exit
rm -f ./bin/refto-cli || exit

echo " "
echo "Server deployed!"
