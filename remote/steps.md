Infrastructure

```bash
terraform init
```
```bash
terraform plan
```
```bash
terraform apply
```

---

Setup scripts

General
```bash
rsync -rPv --delete remote/setup root@<INSTANCE-IP>:/root/remote/
```
```bash
ssh -t root@<INSTANCE-IP> "bash /root/remote/setup/general.sh"
```

Nats-service
```bash
rsync -rPv --delete nats-service/remote/setup root@<INSTANCE-IP>:/root/remote/
```
```bash
ssh -t root@<INSTANCE-IP> "bash /root/remote/setup/nats-service.sh"
```

Proxy-service
```bash
rsync -rPv --delete proxy-service/remote/setup root@<INSTANCE-IP>:/root/remote/
```
```bash
ssh -t root@<INSTANCE-IP> "bash /root/remote/setup/proxy-service.sh"
```

---

Nats-service user
```bash
ssh nats-service@<INSTANCE-IP>
```

Proxy-service user
```bash
ssh proxy-service@<INSTANCE-IP>
```

---

Deploy: Nats-service, Proxy-service
```bash
make remote/deploy
```