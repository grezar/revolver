# Revolver

[![Go Tests](https://github.com/grezar/revolver/actions/workflows/ci.yml/badge.svg)](https://github.com/grezar/revolver/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Overview
Revolver is a CLI-based tool for automating typical key rotation operations written in Go.

You can use YAML to specify the resource from which the key will be issued and the environment in which the key will be used.

```
- name: Example 1
  from:
    provider: AWSIAMUser
    spec:
      accountId: 111
      username: xxx
  to:
    - provider: AWSSharedCredentials
      spec:
        profile: default
    - provider: Tfe
      spec:
        organization: example-org1
        workspace: example-ws1
        secrets:
          - name: AWS_ACCESS_KEY_ID
            value: "{{ .AWSAccessKeyID }}"
            category: "env"
          - name: AWS_SECRET_ACCESS_KEY
            value: "{{ .AWSSecretAccessKey }}"
            category: "env"
```

## Motivation
It is recommended as a best practice from a security perspective that secrets such as IAM User access keys be rotated periodically.
On the other hand, it is very tedious to create a new key,
update it with the new key in the environment using the secret,
and then delete the old key, and in fact the cost of key rotation is not small.
We cannot afford not to automate this tedious but necessary task.
This tool automates key rotation, i.e., it updates the key and automatically updates the secret
so that the newly created key can be used in the environment where the key is used.

One possible scenario is that you are using an AWS IAM User access key in CircleCI or Terraform Cloud, etc.
In this case, the tasks required for key rotation are

1. issue a new AWS IAM User access key
2. update the IAM User access key stored in CircleCI or Terraform Cloud to the newly created one.
3. delete the old AWS IAM User access key

This is what it will look like.
Revolver is a tool to automate exactly this operation.
You can describe the rules for key rotation in a YAML-based configuration and execute the key rotation through the CLI.
By using Revolver, you can automate key rotation, which used to be done by humans, and operate keys safely.

## Usage
Revolver provides a CLI-based interface, so all operations are done through the CLI.

### How to write YAML
The revolver configuration consists of two main sections, from and to, each of which provides a provider that can only be used in that section.
For example, let's say you have an AWS IAM User access key as the target of key rotation, and Terraform Cloud as the environment that uses the User's key.
In this case, we specify the AWSIAMUser provider in the **from** field and the Tfe provider in the **to** field.

Each provider has its own spec, which needs to be set according to the provider.
The YAML in this case is as follows

```
- name: Example 1
  from:
    provider: AWSIAMUser
    spec:
      accountId: abc123
      username: xxx
  to:
    - provider: Tfe
      spec:
        organization: example-org1
        workspace: example-ws1
        secrets:
          - name: AWS_ACCESS_KEY_ID
            value: "{{ .AWSAccessKeyID }}"
            category: "env"
          - name: AWS_SECRET_ACCESS_KEY
            value: "{{ .AWSSecretAccessKey }}"
            category: "env"
```

### Perform key rotations
`revolver rotate` will perform key rotation based on the specified configuration.

You can pass the configuration using `--config` flag so the command to execute will be

```
revolver rotate --config rotations.yaml
```

## Providers
* From
  * [AWSIAMUSer](#from-awsiamuser)

* To
  * [Stdout](#to-stdout)
  * [AWSSharedCredentials](#to-awssharedcredentials)
  * [Tfe](#to-tfe)
  * [CircleCI](#to-circleci)

<a name="from-awsiamuser"></a>
### From/AWSIAMUser

#### Authentication
Since Revolver is using the AWS SDK internally, please refer to the authentication methods available in the AWS SDK

https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/

#### Example
```
  from:
    provider: AWSIAMUser
    spec:
      accountId: abc123
      username: xxx
      expiration: 12h
```

#### Spec
- `accountId` - (Required) AWS Account ID.
- `username` - (Required) AWS IAM User name.
- `expiration` - (Defaults to 90d) Specify the validity period of the key as a string in the following format `1w (week)`, `1d (day)`, `1h (hour)`, `1m (minute)`, `1s (second)`.
   You can also combine them `1w2d3h4m5s`.

#### Secrets
- `.AWSAccessKeyID` - ID of AWS IAM User access key
- `.AWSSecretAccessKey` - Secret key of AWS IAM User access key

<a name="to-stdout"></a>
### To/Stdout
To/Stdout is a provider for outputting something to the stdout

#### Example
```
  to:
    - provider: Stdout
      spec:
        output: |
          say something
```

#### Spec
- `output` - Text to output. You can use Go Template here. Multiple lines are allowed.

<a name="to-awssharedcredentials"></a>
### To/AWSSharedCredentials

#### Authentication
No authentication is required as it only writes to a local file.

#### Example
```
  to:
    - provider: AWSSharedCredentials
      spec:
        profile: example
```

#### Spec
- `path` - (Defaults to ~/.aws/credentials) Path to shared credentials file.
- `profile` - (Defaults to default) AWS Profile name.

<a name="to-tfe"></a>
### To/Tfe
Tfe is for storing secrets provided by *from provider* as Variables in a Workspace hosted by Terraform Cloud/Terraform Enterprise.

#### Authentication
Generate an API token and export it as an environment variable named `REVOLVER_TFE_TOKEN`.

The token needs to be able to list workspaces and read/update workspace variables.

https://www.terraform.io/docs/cloud/users-teams-organizations/api-tokens.html

#### Example
The following is an example of using AWSIAMUser and Tfe in combination.

If you are using AWSIAMUser, you can refer to `.AWSAccessKeyID` and `.AWSSecretAccessKey` using the Go Template in the *to provider* spec.

```
- name: Example 1
  from:
    provider: AWSIAMUser
    spec:
      accountId: 111
      username: xxx
  to:
    - provider: Tfe
      spec:
        organization: org1
        workspace: ws1
        secrets:
          - name: AWS_ACCESS_KEY_ID
            value: "{{ .AWSAccessKeyID }}"
            category: env
          - name: AWS_SECRET_ACCESS_KEY
            value: "{{ .AWSSecretAccessKey }}"
            category: env
```

#### Spec
- `organization` - (Required) Terraform Cloud/Terraform Enterprise Organization name.
- `workspace` - (Required) Terraform Cloud/Terraform Enterprise Workspace name.
- `secrets` - (Required) List of secrets to store.
    - `name` - (Required) Workspace variable name.
    - `value` - (Required) Workspace variable value.
    - `category` - (Defaults to env) Workspace variable category. "env" or "terraform" is available. "env" corresponds to Environment variable, "terraform" corresponds to Terraform variable.

<a name="to-circleci"></a>
### To/CircleCI
CircleCI is a provider for managing secret variables provided by *from provider* in projects or contexts of CircleCI.

#### Authentication
Generate a CircleCI API token and export it as an environment variable named `REVOLVER_CIRCLECI_TOKEN`.

https://circleci.com/docs/2.0/managing-api-tokens/

#### Example
The following is an example of using AWSIAMUser and CircleCI in combination.

If you are using AWSIAMUser, you can refer to `.AWSAccessKeyID` and `.AWSSecretAccessKey` using the Go Template in the *to provider* spec.

```
- name: Example 1
  from:
    provider: AWSIAMUser
    spec:
      accountId: 111
      username: xxx
  to:
    - provider: CircleCI
      spec:
        owner: org1
        projectVariables:
          - project: gh/org1/prj1
            variables:
              - name: AWS_ACCESS_KEY_ID
                value: "{{ .AWSAccessKeyID }}"
              - name: AWS_SECRET_ACCESS_KEY
                value: "{{ .AWSSecretAccessKey }}"
        contexts:
          - name: example-context
            variables:
              - name: AWS_ACCESS_KEY_ID
                value: "{{ .AWSAccessKeyID }}"
              - name: AWS_SECRET_ACCESS_KEY
                value: "{{ .AWSSecretAccessKey }}"
```

#### Spec
- `owner` - (Required) Name of the CircleCI organization.
- `projectVariables` - (Optional) List of the CircleCI project and its variables. Either this or `contexts` is required.
    - `project` - (Required) Name of the CircleCI project.
    - `variables` - (Required) List of the CircleCI project variables to manage.
        - `name` - (Required) Environment variable name of the CircleCI project.
        - `value` - (Required) Environment variable value of the CircleCI project.
- `contexts` - (Optional) List of the CircleCI context and its variables. Either this or `projectVariables` is required.
    - `name` - (Required) Name of the CircleCI context.
    - `variables` - (Required) List of the CircleCI context variables to manage.
      - `name` - (Required) Environment variable name of the CircleCI context.
      - `value` - (Required) Environment variable value of the CircleCI context.

## License
[The MIT License (MIT)](https://https://github.com/grezar/revolver/blob/main/LICENSE)
