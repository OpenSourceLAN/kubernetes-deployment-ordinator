# Kubernetes Deployment Ordinator

Let's be clear from the start - this is a hack. Ideally this would be solved
with StatefulSets or an Operator. But here we are.

Given a deployment yaml file with _n_ `Replicas`, this tool generates _n_
independent deployments with `Replicas=1`, and also updates EnvVars and adds
some labels while it's at it.

Why? Because I wanted to generate some environment variables in an ordered
fashion.

## State

Totally pre-alpha. Hard coded paths, no arguments, just a PoC which I tested
once.

