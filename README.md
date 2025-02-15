# konnect-orchestrator

> :warning: **WARNING: This is a work in progress tool. Do not use in production. The tool is under
heavy development and the API and behavior are subject to change.**

## Setup

* Create or login to a Konnect Organization
* Add a System Account named `konnect-orchestrator`
* Assign the `konnect-orchestrator` account the `Organization Admin` role
* Create a new token for the `konnect-orchestrator` account and store locally
* Initialize your platform repository with the `./koctl init <platform-repo-path>` command
* Copy the sample template files created in the `konnect` folder to setup your organization's full configuration
* Run the `./koctl apply --platform <platform-file> --teams <teams-file> --orgs <orgs-file>` command to apply 
  the configuration to your Konnect organization

### Platform repo requirements

* GH Actions need PR creation privileges (See repo settings)
* Add a `KONNECT_PAT` secret for your GitHub actions workflow with the contents of the token created for the `konnect-orchestrator` account

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
