# Managing Snowflake Applications 

## Table of Content
- [Purpose](#purpose)
- [Background](#background)
- [Principles](#principles)
- [Proposal](#proposal)
    - [Example](#example)
    - [Override mechanism and precedence](#override-mechanism-and-precedence)
        - [Override Indicator](#override-indicator)
        - [Default precedence](#default-precedence)

## Purpose 
This document explores the options to make the boundary between the application team and the platform team more flexible by introducing mechanisms that allow snowflake applications to override the platform-wide configurations. It also retraces the thought process and explains the reasoning behind any decision making.

## Background
In terms of drawing the line between the application teams and the platform team, every organization does things a little differently, and there might be justifications for that. 

What this line represents is the balance between "the extent of organizational standardization need" and "the application team can have enough autonomy to produce business results".

One of Kusion's value propositions is to reduce the cognitive burden for those that don't need to perceive the infrastructure details by means of abstraction. But we have to also recognize the side effect that the more we hide from the application team specification, the less they are able to self-service without relying on the platform team. "Self-service" here refers to not just provisioning resources but also having the ability to customize them to some extent.

So instead of retrofitting every organization into a single model and deny the service for ones that can't (or cross our fingers that everyone buys in on what we believe), we want to explore a possibility where this line isn't set in stone, and more importantly, a paradigm that allows more flexible collaboration and gives the users of KusionStack the autonomy to decide where this line makes sense.

## Principles
- Any shift in the boundary should be a result of discussions between the application and the platform
- In the spirit of lowering developer's cognitive burden, the platform team _should_ manage the majority of the complexities unless explicitly required by application
- This exploration should NOT change the currently proposed collaboration framework

## Proposal
The collaboration design doc proposed a collaboration framework where configuration data for any given Kusion-managed resource should come from 3 sources:
- Platform-owned configurations (`platform.k` in the example)
- Application-SRE-owned configurations (`environment.k` in the example)
- Application-dev-owned configurations (`main.k` in the example)

Among the three sources, platform-owned ones and application-owned ones are backed by different user-facing models (KCL schemas) but they eventually both converge to the same set of specs (including Kubernetes resources and terraform resources).

It's worth mentioning that:
1. The name of the files doesn't matter given they are referenced properly in the stack's `kcl.mod`.
2. `platform.k` may not be a single source but rather a set of files that can be applied at different dimensions of the platform (for example, `prod.k` vs `dev.k`).

This framework highlights the different configurations that should be managed by different roles, by isolating them into two separate, abstracted models that application team and platform each are responsible for filling out.

This approach draws a clear line in terms of separation of concern, by driving different teams into different models that focus on different aspects of the same set of resources. But it doesn't address two questions:
1. What happens when the application team provides a conflicting value with what the platform decides (this could happen because of the `extraConfig` mechanism) and how priority works in that scenario.
2. What happens when this line needs to be shifted around in a snowflake organization or for a snowflake application.

### Example
Let's take a look at a real example where the platform team is responsible for setting the upgrade strategy for an RDS instance and they have decided to set it to "allow automatic major and minor upgrades".

Assuming `platform.k` uses `DatabaseConfiguration` model to define RDS behavior platform-wide:
```
    databaseConfiguration: db.DatabaseConfiguration {
        ...
        awsConfig: {
            auto_minor_version_upgrade: True
            auto_major_version_upgrade: True
            ...
        }
        ...
        advancedConfig: {
            ...
        }
    }
```
`main.k` that application developer authors to provision a database for its application:
```
    database: db.Database {
        engine: "MySQL"
        version: "5.7"
    }
```
`aws.k` that application SRE authors which represents the application-specific configuration in the AWS environment:
```
    database: {
        size: 20
        awsConfig: {
            instanceClass: "db.t3.micro"
            ...
        }
        advancedConfig: {
            ...
        }
    }
```

Now suppose the application team (including the devs and the SREs) decides to choose stability over convenience, and that automatically upgrading major database version is too risky for this application. We want to propose a mechanism where the application team can potentially override this value upon reaching an agreement with the platform team.

A proposed `aws.k` with override:
```
    database: {
        size: 20
        awsConfig: {
            instanceClass: "db.t3.micro"
            ...
        }
        advancedConfig: {
            auto_minor_version_upgrade: False
            auto_major_version_upgrade: False
        }
    }
```

You may notice that the relative path of `auto_minor_version_upgrade` and `auto_major_version_upgrade` is not exactly the same in the `DatabaseConfiguration` model (Under `DatabaseConfiguration.awsConfig`) and the `Database` model (Under `Database.advancedConfig`) - which is fine. The logic to match and merge them is taken care of by the Kusion generator and they should eventually both map to the same terraform attributes in the Kusion spec. Assuming the configuration by application team is taking precedence:
```
- id: hashicorp:aws:aws_db_instance:wordpress
  type: Terraform
  attributes:
    engine: MySQL
    engine_version: "5.7"
    allocated_storage: 20
    instance_class: db.t3.micro
    ...
    auto_minor_version_upgrade: false
    auto_major_version_upgrade: false
  extensions:
    ...
```

### Override mechanism and precedence
We want to introduce a mechanism for overriding the platform configuration that is both safe (to the platform) and convenient (to the application team).

The first assumption we can make here is that platform team should "usually" hold the final say on matters like this most of the time because that's where most of the complexities are. The bottom line is that we want to lower the cognitive burden for the application developers to the point that they shouldn't need to learn about the infrastructure complexities. 

But the complexity doesn't magically disappear, it's only chopped up in smaller pieces and consumed by different parts of the organization - either the application SRE or the platform team, sometimes by Kusion models itself if we provide any out-of-the-box default values.

We are proposing that each configuration entry on the platform side should have an implicit "switch" that determines whether it can be override-able by the application team. This "switch" should be managed on the platform side as an escape hatch for a given configuration entry, rather than something that can be abused by the application if not careful.

#### Override Indicator
There are a few options here:
1. We can either leverage the programmability of KCL as a programming language. For example, a variable starting with an underscore(_) is considered immutable by the KCL interpreter.
2. We can also come up with a designation with the platform schemas and have Kusion generator process it when it detects the pattern. For example, if the convention is "Variables starting with a capital letter is override-able", then the platform can define it as `Auto_major_version_upgrade: True` to grant application team the ability to override it.

Option #1: 
```
    databaseConfiguration: db.DatabaseConfiguration {
        ...
        _auto_minor_version_upgrade: True # Not override-able
        auto_major_version_upgrade: True # Override-able
        ...
    }
```

Option #2: 
```
    databaseConfiguration: db.DatabaseConfiguration {
        ...
        auto_minor_version_upgrade: True # Not override-able
        Auto_major_version_upgrade: True # Override-able
        ...
    }
```

#### Default precedence
As mentioned before, our expectation is that this mechanism to override platform configuration should serve as an escape hatch only in snowflake organizations and/or when managing long-tail applications. The design decisions made here should favor the standardized applications more than snowflakes.

For configuration entries that are not present in the platform configuration, the default behavior is that application team cannot set the value freely.

Take an example for a `backup_retention_period` field that is not present in the platform-side instance of `databaseConfiguration`, and assuming we take option #1.

Omit `backup_retention_period` to default to non-override-able:
```
    databaseConfiguration: db.DatabaseConfiguration {
        ...
        _auto_minor_version_upgrade: True # Not override-able
        auto_major_version_upgrade: True # Override-able
        ...
        extraConfigs: {
            _storage_encrypted: True # Not override-able
        }
    }
```

Explicitly set it to override-able: 
```
    databaseConfiguration: db.DatabaseConfiguration {
        ...
        _auto_minor_version_upgrade: True # Not override-able
        auto_major_version_upgrade: True # Override-able
        ...
        extraConfigs: {
            _storage_encrypted: True # Not override-able
            backup_retention_period: 7 # Override-able
        }
    }
```

Explicitly set it to non-override-able: 
```
    databaseConfiguration: db.DatabaseConfiguration {
        ...
        _auto_minor_version_upgrade: True # Not override-able
        auto_major_version_upgrade: True # Override-able
        ...
        extraConfigs: {
            _storage_encrypted: True # Not override-able
            _backup_retention_period: 7 # Not override-able
            ...
        }
    }
```

## Summary

We can find certain mechanisms that indeed let us provide more flexibility to more freely move the line between the application team and the platform team, but they are also introducing new complexities (such as how to declare a field override-able) into the proposed collaboration framework. However, this is a problem we must solve if we are introducing the `extraConfig` mechanism - which is a `map[string]any` that gets passed through to terraform directly.