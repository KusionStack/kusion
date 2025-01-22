# Kusion Core Workflow Design

## Motivation

Workflow is a sequence of steps involved in moving from the beginning to the end of a working process. [Kusion](https://www.kusionstack.io/docs/) is a modern application delivery and management toolchain. We'd like to define a clear and stable core workflow, to describe how Kusion addresses the difficulty in application delivery. We hope the core workflow to be the user's first step into Kusion. By following the steps defined in the core workflow, user can use Kusion positively and breezily.

The core workflow also guides the R&D iteration of Kusion. Kusion may provide more features (e.g. support more config modules, apply runtimes), or present different product forms (e.g. CLI, SaaS), the core workflow should always be consistent.

## Principle of Workflow Design

We adopt the technology-agnostic view when designing core workflow, that is, focus on the end goal and the steps to achieve the goal, rather than the underlying technologies. Tying workflow to certain technology implementation, underlying software or hardware is what should be avoided. Letting existing technologies limit workflow design is not a good idea. Doing so, a product can adopt new tools to provide ever-improving but still coherent user experience.

When it comes to the actual design work, we take a two-step strategy. Firstly, sketch out the overall. After clarifying the end goal, which should be accomplished at the birth of the product, we usually have a rough idea of how to achieve it. At this point, the end point of the workflow is determined, while the definition of the start point is also crucial, we should give an explicit description. Then the design of overall workflow is accomplished, and can be regarded as the workflow's initial version.

Secondly, clarify the intermediate steps. We hope these steps unambiguous, indispensable, and constitute the optimal path to the end goal. Focusing on what the step achieves is important. Thus, defining an intermediate step turns into defining an intermediate goal, and defining a workflow turns into defining a set of goals. When a definition of workflow's steps is given, we need to evaluate whether these steps are necessary, whether they need to be split or merged. The following principles help us do the estimation:

- Whether the steps are or can be performed by different persons or at different time.
- Whether the execution output of the step is useful, which is usually related to the product form. Take online e-commerce platform for an example. If one adopts the supermarket mode, it needs the shopping cart as a step of its core workflow, the shopping workflow is "browse product, add to shopping cart, place orders". If one adopts the hot seller mode, then the shopping cart is no longer needed, and the shopping workflow is "browse product, place orders".  

## Kusion Core Workflow Design

Kusion focuses on the scenario of the delivery of cloud-native applications, so the core workflow is a delivery operation of the application. In order to solve the problems of high configuration complexity and difficulty in ensuring delivery security, Kusion uses code to describe the application. The start point of Kusion core workflow is application configuration code, and the end point is the actual application resources in the infrastructures.

We divide the Kusion core workflow into three steps:

- **Write**: write application configuration code.
- **Build**: build static application configuration data.
- **Apply**: provision application resources.


<br />![core_workflow](core_workflow.png)

<br />
Kusion is application-centric. Next, we start from the perspective of application developers (including application SREs), to explain what each step plays out, and the necessity and rationality of the steps.

### Step 1: Write
 
In order to reduce the complexity of configuration writing, Kusion uses code to describe the configuration, and provides a concise and universal application configuration model. The model's name is [`AppConfiguration`](https://www.kusionstack.io/docs/concepts/appconfigurations). AppConfiguration simplifies and abstracts underlying resources, shielding configuration items that application developers don't need to pay attention to. Besides, application developers can leverage advantages of code, to avoid repeated and error-prone traditional configuration writing work.

The output of the step *Write* is a piece of application configuration code. The code can fully indicate the application delivery result that the developer focuses on.

### Step 2: Build

The robust running of an application relies on a variety of underlying resources, such as workload, database, load balancer, etc. The step Build renders dynamic application configuration code into static infrastructure resource configuration data, which is described by a unified model `Spec`. Spec is the input of the third step Apply, after format conversion, it will be consumed by underlying infrastructure runtimes.

The output of the step *Build* is a static application configuration data. The data consists of a set of resources, and the resources are what the application exists in the real infrastructures.

At this point, You may have some questions. Kusion aims at reducing the complexity of configuration, and abstracts out the model AppConfiguration. Why is there a step Build? It seems make application developers need to be aware of the underlying infrastructure resources. Will the step increase complexity on the contrary? Can the step be pruned? 

It's true that Spec, i.e. the output of step Build, is more complicated than AppConfiguration. While for the safety reason, step Build should exist in the core workflow. Safety, robustness, and being in line with expected infrastructure changes are super essential for application delivery. We don't want the errors out of nowhere until actual provision, but discover problems in advance and shift risk to the left. Compared to AppConfiguration, Spec is better capable to guarantee the safety. Here are the reasons:

- The code is dynamic, which means that the same configuration code may generate different Specs, due to different execution parameters, building logic, or access to third-party services, thereby provisioning different infrastructures. Spec is static, its stability and closure are much higher than the code. Spec serves as the Source of Truth of configuration data, which contains complete and reliable configuration information. Such division allows the configuration code to focus on reducing the complexity of the configuration writing.
- Executing step Build can detect syntax, semantic and other errors in the configuration code. And as static data, Spec is more convenient for the implementation of policy and audit. It's difficult or even impossible to do it on the code.

In addition, building code into configuration data also reduces the execution load and time of the step Apply, which achieves higher SLO.

A Spec can be regarded as an artifact. Once its safety and reliability is guaranteed, the next step Apply can be carried out. Generally speaking, application developer don't need to read and understand Spec. The successful generation of Spec avoidance of a considerable amount of delivery risks. And there should be other accessories to meet higher security demands, they work on Spec and output developer-friendly security report (Kusion is also working on it, such as PaC). Therefore, the additional cognitive burden to the application developer caused by step Build is negligible, we needn't worry about it.

### Step 3: Apply

Now comes the step of actual provision. For Day-1 delivery, all resources in the Spec will be created. For Day-2 delivery, Kusion obtains the real resources' state in the infrastructure, compared with the Spec to determine whether the resources need to be created, updated, deleted, or remain unchanged, and then perform the corresponding operations.

The output of the step *Apply* is the provision of the application resources in real infrastructures, which are consistent with the expected resources in the Spec. The step *Apply* is the end of Kusion core workflow. If the application developer needs upgrade delivery, start from writing and execute the core workflow once more.

When executing the step Apply, Kusion will first give a preview of the resource changes before real provision. In principle, the preview result needs to be confirmed by the application developer. Although the application configuration code is the abstraction of complex infrastructure resources, which is intended to contain all the configurations the application developer needs to concern about, there are still necessary information that is only exposed during the change process. Especially for the stateful resources, the code change may cause the resource replacement, which may result in the loss of data (you can take database as an example). Such change in data on stateful resources needs to be perceived by the application developer, but is difficult to get described by the application configuration, at least for now. Whether the confirmation of the preview result must be user-aware? And whether a more elegant solution can be given by technical means? We think this is an issue worth discussing, and are constantly conducting R&D, trying to find the silver bullet.

## Summary

This article proposes a design for Kusion core workflow. The workflow contains three steps, which are Write, Build, and Apply, starts from application configuration code, and ends at actual provision, to solve the complexity and security challenge for application delivery. This article firstly illustrates the motivation for designing Kusion core workflow, secondly puts forward the principle of workflow design, and finally gives out the specific Kusion core workflow design and supporting details.
