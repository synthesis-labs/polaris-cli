# Polaris CLI

*Polaris development currently in **pre-release***

The Polaris CLI is a tool designed to make it easier for developers to scaffold and deploy micro-services on a Kubernetes cluster. The Polaris CLI works best with the [Polaris](https://github.com/synthesis-labs/polaris) stack and [other Polaris tools](https://github.com/synthesis-labs?utf8=%E2%9C%93&q=polaris&type=&language=).

# Concepts

## Scaffolds

A scaffold is a set of templates that is used to bootstrap a micro-service app to be modified by the developer and then easily deployed onto a cluster. Scaffolds are contained in repositories. You can use an existing scaffold or create your own.

## Repositories

A repository (or repo) is used to easily manage and source scaffolds. You can use the [Official Polaris Scaffold Repo](https://github.com/synthesis-labs/polaris-scaffolds), use a third party repo or create your own.

# Commands

## Polaris Scaffold

The following commands are used to interact with scaffolds.

### List

Lists available scaffolds in all added repositories.

```
polaris scaffold list
```

### Describe

Provides a description for the named scaffold.

```
polaris scaffold describe <name>
```

Arguments:
```
name (required) - The name of the scaffold
```

### Unpack

Unpacks and deploys a scaffold locally.

```
polaris scaffold unpack <scaffold name> <local name> [--parameters]
```

Arguments:
```
scaffold name (required) - The name of the scaffold
local name (required) - the desired name/path of the local unpacked scaffold
--parameters - parameters used to populate the scaffold template
```

## Polaris Scaffold Repo

These commands are used to interact with repositories containing scaffolds.

### Add

Add the specified repo to the local repo list.

```
polaris scaffold repo add <name> <url> <ref>
```

Arguments:
```
name (required) - the name of the repository
url (required) - the git URL of the repository
ref (required) - the desired git commit reference (branch) of the repository
```

### List

Lists all added repositories.

```
polaris scaffold repo list
```

### Remove

Removes the specified repository if it has been added.

```
polaris scaffold repo remove <name>
```

Arguments:
```
name (required) - the name of the repository to be removed
```

### Update

Performs an update on all added repositories.

```
polaris scaffold repo update
```

Arguments:
```
--force - forces a full refresh (delete and re-download) of all added repositories
```
