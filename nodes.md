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

## Login Issues

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


## IPv6 Tunnels

If your node requires an IPv6 tunnel, such as one setup by `he.net`, you may need some additional help keeping it stable. 

### Handling DHCP changes

Most IPv6 tunnels are only designed for static IP access, but can be reconfigured via a GET URL. Check your documentation, but http://tunnelbroker.net/ for instance uses a URL in the form of:

https://username:hash@ipv4.tunnelbroker.net/nic/update?hostname=id

Add this URL to a secret location, such as ~/.tunnel_url

Then create a script to bring up the tunnel. For instance, I store this in `/root/bin/gif0-tunnel.sh`:

```
#!/bin/sh
#
# he.net tunnel configuration script, appropriate for usage in
# dhcp-exit-hooks. Tested on FreeBSD 12.2-STABLE

set -x -u -e
server_v4="72.52.104.74"
server_v6="2001:470:803d::1"
client_v6="2001:470:803d::2"
prefixlen=128

int_gateway="2001:470:803d:1024::1"
lab_gateway="2001:470:803d:cafe::1"

curl $(cat /root/.tunnel_url)

client_v4=$(ifconfig em2| grep inet | awk '{print $2 }')

ifconfig igb0 inet6 "${int_gateway}" prefixlen 64
ifconfig em0 inet6 "${lab_gateway}" prefixlen 64

ifconfig gif0 destroy
ifconfig gif0 create
ifconfig gif0 tunnel "${client_v4}" "${server_v4}"

ifconfig gif0 inet6 "${client_v6}" "${server_v6}" prefixlen "${prefixlen}"
route -n add -inet6 default "${server_v6}"
ifconfig gif0 up

# this is hacky and shouldn't be necessary
sleep 1
/etc/rc.d/local_unbound restart
/etc/rc.d/rtadvd restart
```

This script can then be trigerred automatically when your DHCP address changes, by creating a `/etc/dhclient-exit-hooks` that calls the script on IP changes. For instance:

```
#!/bin/sh
logger -s /etc/dhclient-exit-hooks has been invoked
logger -s "reason='$reason' if='$interface' med='$medium' new='$new_ip_address' old='$old_ip_address'"

if [ "$reason" = "BOUND" -o "$reason" = "RENEW" -o  "$reason" = "REBOOT" -o "$reason" = "REBIND" ]; then
  logger -s "reason $reason: configuring IPv6 tunnel for ${new_ip_address}"
  /root/bin/gif0-tunnel.sh he
  logger -s "all done here"
elif [ "${new_ip_address}" != "${old_ip_address}" ]; then
  logger -s "new ip: ${new_ip_address} - configuring IPv6 tunnel"
  /root/bin/gif0-tunnel.sh he
  logger -s "all done here"
fi
```

### Restarting the tunnel if offline

It probably isn't necessary, but for extra insurance, I use:

`*/5 * * * * ping6 -c1 www.google.com || /root/bin/gif0-tunnel.sh he`
