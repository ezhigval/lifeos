# GitHub SSH-подключение — настроено (2026-10-04)

## Ключ среды (публичная часть, добавлена в GitHub аккаунт `ezhigval`)
```
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFffsKi9w8shIdmadoThOLbrEmdtz9gbFHjVRCoiEK/T lifeos-agent@workspace
```
Приватный ключ: `~/.ssh/id_ed25519` (вне репозитория, не коммитится).

## Что работает
- `ssh -T git@github.com` → `Hi ezhigval! You've successfully authenticated`
- origin переключён на SSH: `git@github.com:ezhigval/lifeos.git`
  (в ~/.ssh/config прописан IdentityFile для github.com)
- Push ветки выполнен: `qwen-code-b16d2caa-abce-44ec-9723-23f3a3af0838`
  → https://github.com/ezhigval/lifeos/pull/new/qwen-code-b16d2caa-abce-44ec-9723-23f3a3af0838
- Все дальнейшие push/pull — через SSH-ключ, токены больше не нужны.

## Нюансы сети среды
- Прямое подключение к `github.com:22` блокируется — использовать обход:
  `ssh -T git@ssh.github.com -p 443` (работает; git push на port 22 прошёл сам).
- HTTPS к приватному репо без credentials не работает — только SSH.

## ВМ smailikin70@93.77.160.149 — НЕдоступна из среды
Проверено (2026-10-04): ports 22, 2222, 8022, 1022, 2022, 80, 443, 8080, 8443, 30269 — все closed/filtered.
Нужно на стороне ВМ/хостинга:
1. Открыть входящий TCP 22 в firewall/security group **для IPv4** (IP среды может меняться —
   временно разрешить 0.0.0.0/0 или добавить точный IP после определения).
2. Проверить, что sshd слушает (`ss -tlnp | grep :22`) и не фаерволится iptables/nftables.
3. Публичный ключ выше уже добавлен в authorized_keys (если добавлялся).

После открытия порта команда проверки из среды:
```bash
ssh -i ~/.ssh/id_ed25519 smailikin70@93.77.160.149 'echo VM_OK'
```
