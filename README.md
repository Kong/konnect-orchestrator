# konnect-orchestrator

> :warning: **WARNING: This is a work in progress tool. Do not use in production. The tool is under
heavy development and the API and behavior are subject to change.**

## Setup

* Create or login to a Konnect Organization
* Add a System Account named `konnect-orchestrator`
* Assign the `konnect-orchestrator` account the `Organization Admin` role
* Create a new token for the `konnect-orchestrator` account and store locally
* Follow the example in [docs/examples/full.yaml](docs/examples/full.yaml) to
  setup your configuration

## API Publication Logic

* The orchestrator will manage a folder structure in the
  `platform.git` repo. This structure will contain assets
  obtained from managed services, and potentially other generated assets
  so that the Konnect / Kong environments can be managed declaratively
  following a GitOps style process.

```txt
konnect
  kong-financial
    envs
      dev
        teams
          core
            services
              KongFinancial
                user
                  openapi.yaml
```

## Platform repo setup

* GH Actions need PR creation privileges
* The PR workflows are currently manually created in the repo, eventually by the orchestrator
* For now, add a `KONNECT_PAT` from a Konnect PAT

