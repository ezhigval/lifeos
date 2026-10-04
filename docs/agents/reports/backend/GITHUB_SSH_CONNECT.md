# GitHub & VM SSH connectivity (agent environment)

## ⚠️ 2026-10-05: key rotation required
The sandbox was reset between sessions — the previous private key
(`...EK/T`, already added to GitHub account **ezhigval** and to the VM's
`~/.ssh/authorized_keys`) no longer exists in this environment.
A NEW ed25519 keypair was generated here; its public half must be re-added:

```
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIL1fWxrufQ/oWBLSYA7P4igqLLaZx+cGtOac4Tn+Efwn lifeos-agent@workspace
```

- GitHub → Settings → SSH and GPG keys → add as "Authentication" key.
- VM → `echo '<line above>' >> ~/.ssh/authorized_keys`

Old key can be removed from GitHub once the new one works.

## Transport notes
- No system `ssh` binary in sandbox. We provide `/root/.ssh/git-ssh`
  (paramiko-based ssh replacement, also symlinked as `/usr/local/bin/ssh`)
  and `git config --global core.sshCommand "/root/.ssh/git-ssh -o BatchMode=yes"`.
- Outbound TCP to github.com:22 and ssh.github.com:443 works fine.
- Push currently fails with `Authentication failed` only because GitHub does
  not yet know the NEW public key.

## VM 93.77.160.149 (smailikin70) — still blocked at network level
Re-probed 2026-10-05 after user opened port 22 on the VM side:
sshd listening, ufw inactive, authorized_key installed — but TCP connect from
this sandbox to 93.77.160.149:22 STILL times out (no SYN-ACK). Sandbox egress
IP is dynamic Alibaba Cloud range, so a single-IP whitelist will not work;
the drop happens before the VM (Yandex Cloud security group or hoster firewall).

### Verification commands (run on VM)
```bash
sudo journalctl -u ssh --since "10 min ago" --no-pager   # any Connection from 8.x?
sudo iptables -L -n --line-numbers | head -30
nc -zv -w3 <sandbox-ip> 22 2>&1 || true                  # reverse reachability test
```
Preferable fix: Tailscale on the VM + tailnet auth token for the agent.
