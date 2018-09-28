# Kubernetes Deployment Ordinator

**Don't use this** - this was a quick and dirty hack to experiment with some
things. You should use Helm instead. Or consider using a stateful set, with
support inside your container images to dynamically update settings based
on the hostname assigned by the stateful set.

Given a deployment yaml file with _n_ `Replicas`, this tool generates _n_
independent deployments with `Replicas=1`, and also updates EnvVars and adds
some labels while it's at it.

Why? Because I wanted to generate some environment variables in an ordered
fashion.

## State

Totally pre-alpha. Hard coded paths, no arguments, just a PoC which I tested
once.

