## Install
Download the binary from release page, and put it in your PATH. You may need to allow it to execute from the `settings->Security & Privacity` dialog.

## Usages


``` text
List and diff reversions of deployment/daemonset/statefulset

Usage:
  kubectl-v [command]

Available Commands:
  diff        Show diff from different reversions of the resource
  help        Help about any command
  list        list all the reversions of the resource

Flags:
  -h, --help                help for kubectl-v
      --kubeconfig string   Path to the kubeconfig file to use for CLI requests.
  -n, --namespace string    If present, the namespace scope for this CLI request
  -v, --v Level             number for the log level verbosity

Use "kubectl-v [command] --help" for more information about a command.

```

### list command

Use list command to list all the exit reversions: `k v list -n sandbox deploy dong-web`

``` text
 # | CREATE TIME                   | NAME                | DESIRED | AVAILIABE | READY
---+-------------------------------+---------------------+---------+-----------+-------
 2 | 2022-07-13 17:25:14 +0800 CST | dong-web-68dc7787b7 |       2 |         2 |     2
 1 | 2022-07-13 17:24:15 +0800 CST | dong-web-78b7f7cbb  |       0 |         0 |     0
```

Add `-d` to show more details: `k v list -n sandbox deploy dong-web -d`

``` text
 # | CREATE TIME                   | NAME                        | DESIRED | AVAILIABE | READY
---+-------------------------------+-----------------------------+---------+-----------+-------
 2 | 2022-07-13 17:25:14 +0800 CST | dong-web-68dc7787b7         |       2 |         2 |     2
   |                               | └─dong-web-68dc7787b7-6ggpd |         |           |
   |                               | └─dong-web-68dc7787b7-z9dbl |         |           |
 1 | 2022-07-13 17:24:15 +0800 CST | dong-web-78b7f7cbb          |       0 |         0 |     0
```

### diff command

The command below will get the same result
- `k v diff -n sandbox deploy dong-web` show difference between the latest two versions
- `k v diff -n sandbox deploy dong-web 1` show difference between the version 1 and the latest version
- `k v diff -n sandbox deploy dong-web 1 2` show different between version 1 and 2


``` diff
--- OLD
+++ NEW
@@ -4,13 +4,13 @@
   annotations:
     deployment.kubernetes.io/desired-replicas: "2"
     deployment.kubernetes.io/max-replicas: "3"
-    deployment.kubernetes.io/revision: "1"
-  generation: 3
+    deployment.kubernetes.io/revision: "2"
+  generation: 2
```
