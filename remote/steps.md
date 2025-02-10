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

---

Nats-service user
```bash
ssh nats-service@<INSTANCE-IP>
```

---

Deploy
```bash
make remote/deploy
```