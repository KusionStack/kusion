# Declarative Application Configuration Model - AppConfiguration

## Abstract

With the continuous advancement of the cloud-native O&M architecture of the main site and the productization of ATS technology, more and more main site applications and ATS products are evolving towards configuration coding. With the continuous growth of business scenarios and the complexity of the infrastructure configuration involved, the existing application configuration model has become difficult to maintain and unsustainable expansion. Therefore, based on our in-depth analysis of a large number of service models and their dependent infrastructure services, and combined with the latest development trends of industry platform engineering technology, the existing configuration model has been comprehensively upgraded. This article elaborates on the core design concept and design of the next-generation declarative application configuration model. The train of thought and the overall structure of the model are given, which points out the direction for the continuous iteration of the subsequent model.

## Motivation

As the core sub-battlefield of the Red Bull campaign, IaC operation and maintenance designed and built the Kusion technology stack with the core concept of programmability and promoted the evolution of Ant application baseline configuration code. In this process, the first generation of declarative application configuration models - SofaAppConfiguration and SigmaAppConfiguration were implemented. Based on this, the ATS overall production and research delivery process adopts the configuration-driven idea at the beginning of the design, and further broadens the scope of the infrastructure covered by the configuration assets, and precipitates many new configuration models. However, due to the lack of systematic thinking and precipitation in the overall model design, the existing application configuration models cover a wide range, but the design level of the models is not satisfactory, which brings a high cognitive burden to developers. At the same time, due to the lack of standard model design specifications and development guidelines, the research and development of new model requirements is still mainly in charge of the platform team. This is in line with the open collaboration technology promoted by the team Culture goes against. Not forgetting the original intention, reviewing the key starting point when we built the code-based infrastructure and the pits we stepped on along the way in the past, and at the same time looking out at the development trend of platform engineering technology in the industry, improving the developer experience, and empowering developers to self-service has become "to be supplemented".

## Design

### Core Principles

#### Developer First

The AppConfiguration model is a platform interface for end users. The model design should start from the perspective of the platform's end users-developers, rather than platform developers or infrastructure developers. By establishing an abstract description of the application dimension, it can define model attributes with concepts and words that are easy for application developers to understand, shield the complexity of the underlying infrastructure, and reduce the cognitive burden and understanding cost of application developers using the model.

#### Application-Centric

Based on the practical experience of large-scale application delivery in the ATS scenario, from the perspective of the application, the delivery of truly production-available application services not only depends on basic computing resources, but also depends on the network, storage, database, middleware, and monitoring alarms. By defining the abstract mapping of the above-mentioned infrastructure resources, application developers can declare the dependence of the entire life cycle of the application on the infrastructure in a consistent manner, and build a single source of truth centered on the AppConfiguration model, so that a single configuration can fully describe the application.

#### Platform Agnostic

With the diversified development of ant business, the trend of diversification of application technology stacks is becoming more and more obvious. The design of AppConfiguration models is not strong enough to bind any single technology stack. Whether it is Sofa technology stack commonly used in ant or technology stacks based on other languages (Python, Golang, C++, etc.), application developers can describe applications through AppConfiguration models and AppConfiguration built-in multi-cloud support for models with the help of configuration generators. Realize a configuration of multi-cloud migration.

### Model Architecture

By describing all components and various dependent accessories required for application deployment as a unified, infrastructure-independent declarative configuration description, this configuration description serves as an intuitive user intent, thereby achieving standardized and efficient application delivery in a hybrid environment. This enables end users to use rich extension capabilities and self-service operations based on a unified concept without paying attention to the underlying details.![1689919720683-6c6f2afd-ea4f-427c-80ed-3383dd65cbec.png](https://intranetproxy.alipay.com/skylark/lark/0/2023/png/5960/1690786027522-9dfd6b1a-f170-4bbc-9a02-d2feea9a2978.png)
This configuration description consists of five parts, namely components, topology, workflow, policy set, and dependencies. Let's talk about the core positioning and main functions of each part.

### Core Concepts

#### Component

Components are an important part of the entire application, and define the core modules contained in the application that can be actually deployed or the core services that the application depends on. Usually, we think that a complete application configuration description mainly includes a core service for frequent iterations, and a set of middleware collections (including database, cache, cloud service, etc.) that the service depends on. Each component consists of a workload capability (Workload) and any number of accessories (Accessory), which describe the runtime capability and operation and maintenance characteristics that the component depends on. Simply put, we can define a Component with the following expression:
component = workload + accessories
From the perspective of the division of labor and collaboration between different roles, platform R&D and SRE are responsible for defining out-of-the-box components and continuously adding new accessory implementations. Application development and SRE use corresponding components/accessories to describe application operation and maintenance intentions, so as to achieve separation of concerns, so that different roles can be more focused and professional to do their jobs well.

#### Pipeline

Tools and platforms provide a default application delivery process, which can meet most application scenarios. For scenarios that require custom application delivery processes, workflow provides certain scalability. A typical delivery workflow consists of multiple steps, and each step corresponds to the business logic that needs to be executed. Typical workflow steps include manual review, data transfer, multi-cluster release, notification, etc. Specifically at the implementation level, the specific execution of each workflow step is the responsibility of the corresponding plug-in, and platform R&D and SRE can implement new plug-ins.

#### Topology

In order to meet the requirements of high availability and disaster recovery, applications running in the actual production environment usually need to deploy multiple computer rooms and multi-clusters. The deployment topology describes the target computer room or cluster for application delivery and the number of copies in the corresponding environment.

#### PolicySets

The policy set is responsible for defining the technical risks and security compliance policies that need to be followed during the specified application delivery process, such as release policies, risk defense policies, and self-healing policies. Based on standard policy development specifications, platform development, infrastructure SRE, and security students can develop new policy sets, and application development and SRE can bind corresponding policy capabilities based on actual application delivery requirements.

#### Dependency

In a complex business system in production practice, there are often intricate dependencies between applications. Dependency is responsible for defining the dependencies between multiple components within an application and cross-applications.

## References

1. Score - https://docs.score.dev/docs/overview/
2. Acornfile - https://docs.acorn.io/authoring/overview
3. KubeVela - https://kubevela.io/docs/getting-started/core-concept
4. Phoenix Architecture - http://icyfenix.cn/immutable-infrastructure/container/application-centric.html

