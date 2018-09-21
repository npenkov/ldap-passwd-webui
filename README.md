# Web UI for LDAP changing password  

WebUI Client capable of connecting to backend LDAP server and changing the users password.

![Screenshot](screenshots/index.png)

## Running in docker container

```sh
docker run -d -p 8080:8080 --name ldap-passwd-webui \
    -e LPW_TITLE="Change your global password for example.org" \
    -e LPW_HOST="your_ldap_host" \
    -e LPW_PORT="636" \
    -e LPW_ENCRYPTED="true" \
    -e LPW_START_TLS="false" \
    -e LPW_SSL_SKIP_VERIFY="true" \
    -e LPW_USER_DN="uid=%s,ou=people,dc=example,dc=org" \
    -e LPW_USER_BASE="ou=people,dc=example,dc=org" \
    -e LPW_PATTERN='.{8,}' \
    -e LPW_PATTERN_INFO="Password must be at least 8 characters long." \
    npenkov/docker-ldap-passwd-webui:latest
```

## Building and tagging

```sh
make
```

## Credits

 * [Web UI for changing LDAP password - python](https://github.com/jirutka/ldap-passwd-webui)
 * [Gitea](https://github.com/go-gitea/gitea)