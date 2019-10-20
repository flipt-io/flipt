# Integration

This document describes how to integrate Flipt in your existing applications. To learn how to install and run Flipt, see the [Installation](installation.md) documentation.

Once you have the Flipt server up and running within your infrastructure, the next step is to integrate the Flipt client(s) with your applications that you would like to be able to use with Flipt.

There are two ways to communicate with the Flipt server:

1. GRPC
1. REST API

## Flipt GRPC Clients

Since Flipt is a [GRPC](https://grpc.io/) enabled application (see: [Architecture](architecture.md)), to communicate with the Flipt server, you can use a generated GRPC client for your language of choice.

This means that your application can use the Flipt GRPC client if it is written in one of the many languages that GRPC supports, including:

* C++
* Java
* Python
* Go
* Ruby
* C#
* Node.js
* Android Java
* Objective-C
* PHP

The Flipt GRPC client is the preferred way to integrate your application with Flipt as it is more performant than REST and requires the least amount of configuration.

An example Go application exists at [https://github.com/markphelps/flipt/examples/basic](https://github.com/markphelps/flipt/tree/master/examples/basic), showing how you would integrate with Flipt using the Go GRPC client.

### Download

Prebuilt Flipt GRPC clients are currently available for the following languages:

* Go: [https://github.com/markphelps/flipt-grpc-go](https://github.com/markphelps/flipt-grpc-go)
* Ruby: [https://github.com/markphelps/flipt-grpc-ruby](https://github.com/markphelps/flipt-grpc-ruby)

If your language is not listed, please see the section below on how to generate a native GRPC client manually. If you choose to open source this client, please submit a pull request so I can add it to the docs.

### Generate

If a GRPC client in your language is not available for download, you can easily generate it yourself using the existing [protobuf definition](https://github.com/markphelps/flipt/blob/master/rpc/flipt.proto). The [GRPC documentation](https://grpc.io/docs/) has extensive examples on how to generate GRPC clients in each supported language.

!!! note
    GRPC generates both client implementation and the server interfaces. To use Flipt you only need the GRPC client implementation and can ignore the server code as this is implemented by Flipt itself.

Below are two examples on how to generate Flipt clients in both Go and Ruby.

#### Go Example

1. Follow setup here: [https://grpc.io/docs/quickstart/go/](https://grpc.io/docs/quickstart/go/)
2. Generate using protoc to desired location:

```bash
$ protoc -I ./rpc --go_out=plugins=grpc:/tmp/flipt/go ./rpc/flipt.proto
$ cd /tmp/flipt/go/flipt
$ ls
flipt.pb.go          flipt_pb.rb          flipt_services_pb.
```

#### Ruby Example

1. Follow setup here: [https://grpc.io/docs/quickstart/ruby/](https://grpc.io/docs/quickstart/ruby/)
2. Generate using protoc to desired location:

```bash
$ grpc_tools_ruby_protoc -I ./rpc --ruby_out=/tmp/flipt/ruby --grpc_out=/tmp/flipt/ruby ./rpc/flipt.proto
$ cd /tmp/flipt/ruby
$ ls
flipt_pb.rb          flipt_services_pb.rb
```

## Flipt REST API

Flipt also comes equipped with a fully functional REST API. In fact, the Flipt UI is completely backed by this same API. This means that anything that can be done in the Flipt UI can also be done via the REST API.

The Flipt REST API can also be used with any language that can make HTTP requests. This means that you don't need to use one of the above GRPC clients in order to integrate your application with Flipt.

The latest version of the REST API is fully documented using OpenAPI v2 (formerly Swagger) specification available [here](https://github.com/markphelps/flipt/blob/master/swagger/api/swagger.json).

Each Flipt server instance also hosts it's own REST API documentation. This documentation is reachable in the Flipt UI by clicking the **API** link in the header navigation.

![Flipt API](assets/images/integration/api.png)

This will load the API documentation which documents valid requests/responses to the Flipt REST API:

![Flipt API Docs](assets/images/integration/docs.png)

## Flipt REST Clients

### Generate

You can use [swagger-codegen](https://github.com/swagger-api/swagger-codegen) to generate client code in your prefered language from the OpenAPI v2 specification linked above.

While generating clients in all languages supported by swagger-codegen is outside of the scope of this documentation, an example of generating a Java client is below.

#### Java Example

1. Install `swagger-codegen`: [https://github.com/swagger-api/swagger-codegen#prerequisites](https://github.com/swagger-api/swagger-codegen#prerequisites)
1. Generate using `swagger-codegen` to desired location:

```bash
swagger-codegen generate -i swagger/api/swagger.json -l java -o /tmp/flipt/java
```

## Third-Party Client Libraries

Client libraries built by awesome people from the Open Source community:

* [Camji55/Flipt-iOS-SDK](https://github.com/Camji55/Flipt-iOS-SDK) - Native iOS SDK for Flipt (Swift)
* [christopherdiehl/rflipt](https://github.com/christopherdiehl/rflipt) - React components/example project to control React features backed by Flipt (React)
