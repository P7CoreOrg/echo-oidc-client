# echo-oidc-client

## Developing

This application is backed by the Echo framework. To get started with development, you'll need to restore the build locally.

```bash
go mod download
```
```bash
<root>\pkg\P7CoreOrg\go-oidc>git submodule update
```



Once all dependencies are installed, you can run the CLI easily.

```bash
go run .\afx
go run .\afx login
go run .\afx login -s
```


afx is a native cli that does an OIDC login.  It does this by hosting a temporary http server which gets an authorization code via the IDP redirect.  





