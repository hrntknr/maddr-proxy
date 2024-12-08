# maddr proxy

## Usage

### Server

```
LISTEN=:1080 PASSWORD=hoge,huga ./maddr-proxy
```

environment variables:

| name     | default |
| -------- | ------- |
| LISTEN   | :1080   |
| PASSWORD |         |

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
