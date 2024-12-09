# maddr proxy

This is an implementation of a proxy server with selectable source address.  
Use interface name or ip address for the proxy user name.

## Usage

```
Usage:
  maddr-proxy proxy [flags]

Flags:
  -h, --help                       help for proxy
  -l, --listen string              listen address (default ":1080")
  -p, --password string            password
      --setup-route                setup route
      --setup-route-iface string   interface match (default "en.*,eth.*")
```

```
Usage:
  maddr-proxy setup-route [flags]

Flags:
  -h, --help           help for setup-route
  -i, --iface string   interface match (default "en.*,eth.*")
  -w, --watch          watch
```

Using setup-route mode or --setup-route option installs a policy based routing route.  
The route to be installed is as follows

```sh
hrntknr@proxy1:~$ ip a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host noprefixroute
       valid_lft forever preferred_lft forever
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP group default qlen 1000
    link/ether 60:45:bd:67:b5:c1 brd ff:ff:ff:ff:ff:ff
    inet 10.64.0.5/24 metric 100 brd 10.64.0.255 scope global eth0
       valid_lft forever preferred_lft forever
    inet6 fe80::6245:bdff:fe67:b5c1/64 scope link
       valid_lft forever preferred_lft forever
3: eth1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP group default qlen 1000
    link/ether 60:45:bd:67:bb:88 brd ff:ff:ff:ff:ff:ff
    inet 10.64.0.4/24 metric 200 brd 10.64.0.255 scope global eth1
       valid_lft forever preferred_lft forever
    inet6 fe80::6245:bdff:fe67:bb88/64 scope link
       valid_lft forever preferred_lft forever
4: docker0: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc noqueue state DOWN group default
    link/ether 02:42:e4:6c:81:dd brd ff:ff:ff:ff:ff:ff
    inet 172.17.0.1/16 brd 172.17.255.255 scope global docker0
       valid_lft forever preferred_lft forever
hrntknr@proxy1:~$ ip rule
0:      from all lookup local
15100:  from 10.64.0.4 lookup 15100
32766:  from all lookup main
32767:  from all lookup default
hrntknr@proxy1:~$ ip route show table 15100
default via 10.64.0.1 dev eth1 proto 151
```

### Client

```sh
# normal request
curl https://ifconfig.io/ -x http://localhost:1080

# request with interface name
curl https://ifconfig.io/ -x http://ens3:@localhost:1080

# request with network and interface name (only support tcp4/tcp6)
curl https://ifconfig.io/ -x http://tcp4:ens3:@localhost:1080
curl https://ifconfig.io/ -x http://tcp6:ens3:@localhost:1080

# request with address
curl https://ifconfig.io/ -x http://10.0.0.2:@localhost:1080
curl https://ifconfig.io/ -x http://2001:0db8::3456:::@localhost:1080

# request with auth
curl https://ifconfig.io/ -x http://:password@localhost:1080

# request with interface name and password
curl https://ifconfig.io/ -x http://ens3:password@localhost:1080

# request with network and interface name and password
curl https://ifconfig.io/ -x http://tcp4:ens3:password@localhost:1080
curl https://ifconfig.io/ -x http://tcp6:ens3:password@localhost:1080

# request with address and password
curl https://ifconfig.io/ -x http://10.0.0.2:password@localhost:1080
curl https://ifconfig.io/ -x http://2001:0db8::3456:::password@localhost:1080
```
