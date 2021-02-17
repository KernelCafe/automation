# Node Documentation

## Security Advice

### Firewall Recommendations

It is highly recommended to have a firewall between cafe nodes and any other networks, including the Internet.

The only incoming traffic necessary is port 22 (sshd) via IPv6. 

#### Example

Here is an example `pf.conf` configuration with an IPv6 6-in-4 tunnel:

```
ext_if="em2"
int_if="igb0"
cafe_if="em0"
tunnel_if="gif0"
cafe_v6="2001:470:803d:cafe::1/64"

tcp_services="{ 22 }"

icmp_types="echoreq"
icmp6_types = "{ echoreq, unreach, timex, paramprob, redir, routeradv, routersol }"

set block-policy return
set loginterface $ext_if
set skip on lo
scrub in on $ext_if
scrub in on $cafe_if

nat on $ext_if inet from !($ext_if) -> ($ext_if:0) static-port
block in log
pass out keep state
antispoof quick for { lo $int_if $cafe_if }

pass in log on $tunnel_if inet6 proto tcp from any to $cafe_v6 port $tcp_services
pass in log inet proto icmp all icmp-type $icmp_types
pass in log inet6 proto icmp6 all icmp6-type $icmp6_types

pass quick on $int_if
pass quick on $cafe_if

pass out quick on $tunnel_if inet6 from any to any
```

### sshd

To prevent unauthorized access, it is recommended to disable unnecessary authentication mechanisms in sshd, with the following `/etc/ssh/sshd_config` settings:

```
PubkeyAuthentication yes
PasswordAuthentication no
HostbasedAuthentication no
ChallengeResponseAuthentication no
KerberosAuthentication no
```

### Login Issues

### macOS

If users cannot login, set `UsePAM=false` in `/etc/ssh/sshd_config`. macOS otherwise does not allow access to users without explicit passwords set.

### Fedora

If users cannot login, check for SELinux errors via `journalctl`. If you see entries that resemble:

```
Feb 16 09:43:26 shrimp-paste audit[311151]: USER_AUTH pid=311151 uid=0 auid=4294967295 ses=4294967295 subj=system_u:system_r:sshd_t:s0-s0:c0.c1023 msg='op=pubkey acct="t" exe="/usr/sbin/sshd" hostname=? addr=2001:470:803d:1024:cc29:f36c:2483:dd26 terminal=ssh res=failed'
Feb 16 09:50:41 shrimp-paste audit[311891]: AVC avc:  denied  { read } for  pid=311891 comm="sshd" name="authorized_keys" dev="nvme0n1p2" ino=27263159 scontext=system_u:system_r:sshd_t:s0-s0:c0.c1023 tcontext=system_u:object_r:default_t:s
0 tclass=file permissive=0 
```

This is because on Fedora, SELinux is configured to only allow reading `authorized_keys` from `/home/`, and kernel.cafe users live in `/u/`. You can corfirm by checking if `ls -Zd /u/t` reports `system_u:object_r:etc_t:s0`, rather than `unconfined_u:object_r:user_home_dir_t:s0`


1. Run: `sudo vim /etc/selinux/semanage.conf`
2. Set `usepasswd=true`
3. Run: `sudo  restorecon -Rv /u`
