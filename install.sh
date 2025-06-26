#!/bin/bash

# echo "export PDNS_KEY=a7eadf75278dd026e54c24d3aeff992cd5a8fc19" >> ~/.bashrc

# source ~/.bashrc

go build .

mkdir -p usr/local/bin

cp ./conductor /usr/local/bin/conductor

chmod +x /usr/local/bin/conductor

mkdir -p ~/.config/systemd/user

service="[Unit]
Description=Conductor Service
After=network.target redis.service

[Service]
ExecStart=/usr/local/bin/conductor
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
