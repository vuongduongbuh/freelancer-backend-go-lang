#!/bin/bash

if [[ $1 = "dev" ]]; then
    HOST='dev.nowatwork.ch'
elif [[ $1 = "prod" ]]; then
    HOST='wrk01.nowatwork.ch'
else
    echo "unknown env"
    exit 1
fi

if [[ $1 != "dev" ]]
then
    read -p "DANGER: do you want to deploy to $1 ($HOST)? " -n 1 -r
    echo    # (optional) move to a new line
    if [[ ! $REPLY =~ ^[Yy]$ ]]
    then
        exit 1
    fi
fi

echo
echo "deployment to $HOST started..."
echo

SERVICE='nowatwork-srv'

gox -osarch="linux/amd64" -output "build/"$SERVICE"_linux_amd64"
cp main.env build/main.env
cp prod.env build/prod.env
cp process.json build/process.json
cp -R assets build/assets
rm build/assets/images/*

if [[ $1 = "dev" ]]
then
    rm build/process.json
    cp dev-process.json build/process.json
fi

rsync -az --force --progress -e "ssh -p22" ./build/* root@$HOST:/var/projects/$SERVICE/
rm -r build
