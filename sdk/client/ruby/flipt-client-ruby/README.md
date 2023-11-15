## Flipt Client Ruby

The `flipt-client-ruby` directory contains the Ruby source code for a Flipt evaluation client using FFI to make calls to a core built in Rust.

### Instructions

To use this client, you can run the following command from the root of the repository:

```bash
cargo build --release
```

This should generate a `target/` directory in the root of this repository, which contains the dynamically linked library built for your platform. This dynamic library will contain the functionality necessary for the Ruby client to make FFI calls. 

TODO: Currently, you'll need copy the dynamic library to the `flipt-client-ruby/lib/ext` directory. This is a temporary solution until we can figure out a better way to package the libraries with the gem.

The `path/to/lib` will be the path to the dynamic library which will have the following paths depending on your platform.

- **Linux**: `{FLIPT_REPO_ROOT}/target/release/libfliptengine.so`
- **Windows**: `{FLIPT_REPO_ROOT}/target/release/libfliptengine.dll`
- **MacOS**: `{FLIPT_REPO_ROOT}/target/release/libfliptengine.dylib`

You can then build the gem and install it locally:

```bash
rake build
gem install pkg/flipt_client-0.1.0.gem
```

In your Ruby code you can import this client and use it as so:

```ruby
require 'flipt_client'

client = Flipt::EvaluationClient.new() # uses 'default' namespace
resp = client.variant({ flagKey: 'buzz', entityId: 'someentity', context: { fizz: 'buzz' } })

puts resp
```
