#!/bin/bash

service_exists() {
 systemctl --user list-unit-files --type=service --all | grep -Fq "conductor.service"
}

service_is_active() {
 systemctl --user is-active --quiet "conductor.service"
}

if service_exists; then
 if service_is_active; then
  systemctl --user stop "conductor.service"
 fi
fi

go build .

mkdir -p ~/bin

cp ./conductor ~/bin/conductor

chmod +x ~/bin/conductor

mkdir -p ~/.config/systemd/user

service="[Unit]
Description=Conductor Service
After=network.target

[Service]
ExecStart=/home/$(whoami)/bin/conductor
Restart=always
Environment=PDNS_KEY=a7eadf75278dd026e54c24d3aeff992cd5a8fc19

[Install]
WantedBy=default.target
"

echo "$service" | tee ~/.config/systemd/user/conductor.service > /dev/null

systemctl --user daemon-reexec

systemctl --user daemon-reload

systemctl --user enable conductor

systemctl --user start conductor
