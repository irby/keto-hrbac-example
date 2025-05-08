# Example Ory Hydra HRBAC Implementation

## Requirements
- Docker ([link](https://www.docker.com/get-started/))
- Docker Compose ([link](https://docs.docker.com/compose/))

## Getting Started
If you have Docker and Docker Compose installed on your machine, you should be ready to get started with this project stack.

To spin up the Docker containers required for this project, simply run the following in your terminal:

```bash
docker-compose up --build
```

Alternatively, if `docker-compose` is not recognized on your machine you can try the following:

```bash
docker compose up --build
```

### Docker Containers Being Run:

#### Keto

The `keto` container is the erver that runs [Ory Keto](https://www.ory.sh/keto/), an open-source, low-latency RBAC system that dictates authorzation of [subjects](https://www.ory.sh/docs/keto/concepts/subjects) to [objects](https://www.ory.sh/docs/keto/concepts/objects) within [namespaces](https://www.ory.sh/docs/keto/concepts/namespaces). Within this implementation, everything is handled in-memory, so added relationships are not persisted between runs. However, Ory Keto can be configured to persist information to a database.

The configuration for the Keto instance can be found within [./keto/config/keto.yml](./keto/config/keto.yml). We configure the Keto instance with the following:

- The read/write servers listen on ports 4467/4468, respectively.
- The server is configured with two namespaces:
    - document
    - role

Namespaces are static constructs in Ory Keto, so they cannot be modified outside of the Keto configuration.

The container serves a [REST API](https://www.ory.sh/docs/keto/reference/rest-api/) and [gRPC API](https://www.ory.sh/docs/keto/reference/proto-api/). The example Go application will make calls to the Keto instance using the gRPC API.

#### Keto-Init

The `keto-init` container is responsible for initializing the Keto server with some starter [relationship tuples](https://www.ory.sh/docs/keto/concepts/relation-tuples) that will be used for the example application. The relation tuples can be found within the [./keto/relation-tuples](./keto/relation-tuples/) folder.

When the container is run, it will push each payload in this directory to the `keto` container.

Here is a breakdown of the relation-tuple documents:

##### `admin_editor_relation.json`
This payload specifies that any member of the `admin` role inherits the `editor` role.

##### `editor_viewer_relation.json`
This payload specifies that any member of the `editor` role inherits the `viewer` role.

##### `document_editor_role.json`
This payload specifies that members of the `editor` role can perform the `edit` action role on `document` object `480158d4-0031-4412-9453-1bb0cdf76104`.

##### `document_viewer_role.json`
This payload specifies that members of the `viewer` role can perform the `view` action role on `document` object `480158d4-0031-4412-9453-1bb0cdf76104`.

Noticably, there is no specific payload involving an admin relationship to the document. This omission was intentional. The addition of this relationship will be handled by the example Go application.

#### Example App

TODO

### Hierarchical Role-Based Access Control

With the payloads created by `keto-init` and the example Go application, we can see some hierarchical role-based access control take effect.

In particular, we can see that although we do not explicitly say "admin role has view access to document 480158d4-0031-4412-9453-1bb0cdf76104" from our API definitions, the Keto server is able to determine the relationship between the `admin` and `viewer` roles by proxy of the `editor` role.

This hierarchical mapping of access is extremely powerful and flexible. It reduces complexity and storage, and takes the burden of access control off of the application.

This repository only serves as an example structure of HRBAC, but real-world applications can be realized with more advanced structures.