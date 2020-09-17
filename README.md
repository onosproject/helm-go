# Helm Go Client

The Helm Go Client is a fluent client library for managing [Helm] charts and operating on resources deployed 
by Helm in [Kubernetes]. 

The client has a low barrier to entry for existing `helm` users. The API is modeled on the `helm` CLI commands. 
Additionally, a custom Kubernetes client supports querying and filtering resources by specific Helm chart releases.

## Installation

Helm-Go uses [Go modules](https://golang.org/ref/mod) for dependency management. To add the client to your go module:

```bash
> GO111MODULE=on go get github.com/onosproject/helm-go
```

## Usage

The primary API for the client is the `Helm` interface, which provides functions for setting up repositories, 
querying charts, managing chart releases, and operating on a release's resources in Kubernetes.

To create a `Helm` client, simply call `helm.New()`:

```go
import "github.com/onosproject/helm-go/pkg/helm"

var client = helm.New()
```

By default, the client will be configured with the `default` namespace. You can optionally specify a different
namespace within which to operate:

```go
import "github.com/onosproject/helm-go/pkg/helm"

var client = helm.New("onos")
```

The client exposes a series of sub-client interfaces which replicate the functionality of Helm CLI commands. Interfaces
follow a common pattern, providing fluent builders for requests to the Kubernetes cluster. Requests are always executed
by calling the `Do()` method.

### Managing Repositories

The repository client wraps the functionality provided by the `helm repo` subcommands in the Helm CLI. Repositories
can be added and removed from the repo registry. The repo client can be accessed by calling `Repos()` on the `Helm` 
client:
                                                 
```go
repos := client.Repos()
```


To add a repository to the registry, execute an `Add` request:

```go
repo, err := client.Repos().
	Add("onos").
	URL("https://onos-helm-charts.onosproject.org").
	Username("jordan").
	Passsword("pA$$w0®D").
	Do()
```

To remove a repository, execute a `Remove` request:

```go
err := client.Repos().
	Remove("onos").
	Do()
```

To list the repositories in the registry, call `List`:

```go
repos, err := client.Repos().List()
```

To get a specific repository, use `Get`:

```go
repo, err := client.Repos().Get("onos")
```

### Accessing Charts

The chart client provides an interface similar to the repo client above, allowing users to access information regarding
specific charts. The chart client can be accessed by calling `Charts()` on the `Helm` client:

```go
charts := client.Charts()
```

To load a chart, use `Get`:

```go
chart, err := client.Charts().Get("onos/onos-classic")
```

### Managing Releases

The release client can be used to execute commands for installing, uninstalling, upgrading, etc your Helm charts.
The release client is accessed by calling `Releases()` on the `Helm` client:

```go
releases := client.Releases()
```

To release a chart, execute an `Install` request:

```go
release, err := client.Releases().
	Install("onos", "onos/onos-classic").
	Version("2.5.0").
	Do()
```

If the repository has not been added to the repo registry, repository settings can be provided within the install request:

```go
release, err := client.Releases().
	Install("onos", "onos-classic").
	Repo("https://onos-helm-charts.onosproject.org").
	Username("jordan").
	Passsword("pA$$w0®D").
	Version("2.5.0").
	Do()
```

The `Install` request supports all the same options available through the Helm CLI. For example, you can `Wait` for
the release to complete, set a `Timeout`, perform a `DryRun`, or enable all the other options available when installing
a chart:

```go
release, err := client.Releases().
	Install("onos", "onos/onos-classic").
	DryRun().
	Do()

release, err = client.Releases().
	Install("onos", "onos/onos-classic").
	Version("2.5.0").
	Wait().
	Timeout(5*time.Minute).
	Do()
```

The `Set` method mimics the format of the `--set` flag in the Helm CLI:

```go
release, err := client.Releases().
	Install("onos", "onos/onos-classic").
	Version("2.5.0").
	Set("replicas", 3).
	Set("atomix.replicas", 3).
	Set("heap", "8G").
	Set("apps", []string{"org.onosproject.openflow", "org.onosproject.p4runtime"}).
	Do()
```

Values set using the `Set` method will override the default chart values. Nested values can be set using the same
`dot.notation` used in the Helm CLI. Values can be of a scalar type, map, slice, or struct.

To uninstall a chart release, execute an `Uninstall` request:

```go
err := client.Releases().
	Uninstall("onos").
	Do()
```

### Querying Resources

Once a chart has been installed, the `Release` provides a release-scoped Kubernetes client for querying chart objects.
The release-scoped client can be used to list resources created by the release. This can be helpful for writing
integration tests:

```go
// Get a list of pods created by the atomix-controller
pods, err := release.Client().CoreV1().Pods().List()
assert.NoError(t, err)

// Get the Atomix controller pod
pod := pods[0]

// Delete the pod
err := pod.Delete()
assert.NoError(t, err)
```

Additionally, Kubernetes objects that create and own other Kubernetes resources -- like `Deployment`, `StatefulSet`, 
`Job`, etc -- provide scoped clients that can be used to query the resources they own as well:

```go
// Get the atomix-controller deployment
deps, err := release.Client().AppsV1().Deployments().List()
assert.NoError(t, err)
assert.Len(t, deps, 1)
dep := deps[0]

// Get the pods created by the controller deployment
pods, err := dep.CoreV1().Pods().List()
assert.NoError(t, err)
assert.Len(t, pods, 1)
pod := pods[0]

// Delete the controller pod
err = pod.Delete()
assert.NoError(t, err)

// Wait a minute for the controller deployment to recover
err = dep.Wait(1 * time.Minute)
assert.NoError(t, err)

// Verify the pod was recovered
pods, err := dep.CoreV1().Pods().List()
assert.NoError(t, err)
assert.Len(t, pods, 1)
assert.NotEqual(t, pod.Name, pods[0].Name)
```

[Helm]: https://helm.sh/
[Kubernetes]: https://kubernetes.io/
