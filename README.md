## Conductor service

This is a user level systemd service which has no root privileges, It is because this service is a part of the system which directly interacts with the user via websocket. This service is responsible for managing the playground, proxying websocket and helps running examiner tasks by passing the requests to the examiner binary via REST API.

This service can be installed on a system using the install script provided:

```bash
chmod +x ./install.sh
./install.sh
```

You can check the status by using the following command:
```bash
systemctl --user status conductor
```

Stop service:
```bash
systemctl --user stop conductor
```

Start and restart service:
```bash
systemctl --user start conductor
systemctl --user restart conductor
```

The service exposes a port 8082 which accepts requests from the frontend