Golang Vanity URLs
---------------

Golang Vanity URLs uses AWS API Gateway to enable custom domains for your Go packages.

This package is based on [https://github.com/GoogleCloudPlatform/govanityurls](https://github.com/GoogleCloudPlatform/govanityurls)
and adapts it to work with AWS API Gateway.

### Quickstart

Install [up](https://github.com/apex/up).  The simplest method is:

```bash
curl -sf https://up.apex.sh/install | sh
```

Register your customer domain using [Route 53](https://aws.amazon.com/route53/).

Edit `vanity.yml` and add your custom repos. For example, to use the custom domain, 
`example.com`, to host the `swag` package, I would write: 

```yaml
host: example.com
paths: 
  /swag:
    repo: https://github.com/savaki/swag
```

Deploy the app using `up`:

```bash
up
```

Generate a custom TLS certificate using [AWS Certificate Manager](https://aws.amazon.com/certificate-manager/).

Create a CloudFront distribution for your app.  Use the certificate you create via the AWS Certificate Manager 
to secure the site. To get the URL to point to, type:

```bash
up url
```

That's it!  You can use `go get` to get the packages from your custom domain.


## Configuration File

```
host: example.com
max_age: 3600
paths:
  /foo:
    repo: https://github.com/example/foo
    display: "https://github.com/example/foo https://github.com/example/foo/tree/master{/dir} https://github.com/example/foo/blob/master{/dir}/{file}#L{line}"
    vcs: git
```

<table>
  <thead>
    <tr>
      <th scope="col">Key</th>
      <th scope="col">Required</th>
      <th scope="col">Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <th scope="row"><code>max_age</code></th>
      <td>optional</td>
      <td>The amount of time to cache package pages in seconds.  Controls the <code>max-age</code> directive sent in the <a href="https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control"><code>Cache-Control</code></a> HTTP header.</td>
    </tr>
    <tr>
      <th scope="row"><code>host</code></th>
      <td>optional</td>
      <td>Host name to use in meta tags.  If omitted, uses the App Engine default version host or the Host header on non-App Engine Standard environments.  You can use this option to fix the host when using this service behind a reverse proxy or a <a href="https://cloud.google.com/appengine/docs/standard/go/how-requests-are-routed#routing_with_a_dispatch_file">custom dispatch file</a>.</td>
    </tr>
    <tr>
      <th scope="row"><code>paths</code></th>
      <td>required</td>
      <td>Map of paths to path configurations.  Each key is a path that will point to the root of a repository hosted elsewhere.  The fields are documented in the Path Configuration section below.</td>
    </tr>
  </tbody>
</table>

### Path Configuration

<table>
  <thead>
    <tr>
      <th scope="col">Key</th>
      <th scope="col">Required</th>
      <th scope="col">Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <th scope="row"><code>display</code></th>
      <td>optional</td>
      <td>The last three fields of the <a href="https://github.com/golang/gddo/wiki/Source-Code-Links"><code>go-source</code> meta tag</a>.  If omitted, it is inferred from the code hosting service if possible.</td>
    </tr>
    <tr>
      <th scope="row"><code>repo</code></th>
      <td>required</td>
      <td>Root URL of the repository as it would appear in <a href="https://golang.org/cmd/go/#hdr-Remote_import_paths"><code>go-import</code> meta tag</a>.</td>
    </tr>
    <tr>
      <th scope="row"><code>vcs</code></th>
      <td>required if ambiguous</td>
      <td>If the version control system cannot be inferred (e.g. for Bitbucket or a custom domain), then this specifies the version control system as it would appear in <a href="https://golang.org/cmd/go/#hdr-Remote_import_paths"><code>go-import</code> meta tag</a>.  This can be one of <code>git</code>, <code>hg</code>, <code>svn</code>, or <code>bzr</code>.</td>
    </tr>
  </tbody>
</table>