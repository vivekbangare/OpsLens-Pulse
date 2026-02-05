# Linux Installation

## Install Packages
```bash
sudo dpkg -i opslens-pulse-server_amd64.deb
sudo dpkg -i opslens-pulse-agent_amd64.deb
```

Enable Services

```
sudo systemctl daemon-reload
sudo systemctl enable opslens-pulse-server
sudo systemctl start opslens-pulse-server

sudo systemctl enable opslens-pulse-agent
sudo systemctl start opslens-pulse-agent

```

Logs

```
/var/log/opslens-pulse/server.log

/var/log/opslens-pulse/agent.log
```