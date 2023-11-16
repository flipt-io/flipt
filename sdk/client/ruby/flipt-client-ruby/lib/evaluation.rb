require "client/version"
require "ffi"

module Flipt
  class Error < StandardError; end

  class EvaluationClient
    FLIPTENGINE="ext/libfliptengine"

    def self.platform_specific_lib
      case RbConfig::CONFIG['host_os']
      when /darwin|mac os/
        "#{FLIPTENGINE}.dylib"
      when /linux/
        "#{FLIPTENGINE}.so"
      when /mswin|msys|mingw|cygwin|bccwin|wince|emc/
        "#{FLIPTENGINE}.dll"
      else
        raise "unsupported platform #{RbConfig::CONFIG['host_os']}"
      end
    end

    extend FFI::Library
    ffi_lib File.expand_path(platform_specific_lib, __dir__)

    # void *initialize_engine(const char *const *namespaces);
    attach_function :initialize_engine, [:pointer], :pointer
    # void destroy_engine(void *engine_ptr);
    attach_function :destroy_engine, [:pointer], :void
    # const char *variant(void *engine_ptr, const char *evaluation_request);
    attach_function :variant, [:pointer, :string], :string
    # const char *boolean(void *engine_ptr, const char *evaluation_request);
    attach_function :boolean, [:pointer, :string], :string

    def initialize(namespace = "default")
      @namespace = namespace
      namespace_list = [namespace]
      ns = FFI::MemoryPointer.new(:pointer, namespace_list.size)
      namespace_list.each_with_index do |namespace, i|
        ns[i].put_pointer(0, FFI::MemoryPointer.from_string(namespace))
      end
      @engine = self.class.initialize_engine(ns)
      ObjectSpace.define_finalizer(self, self.class.finalize(@engine))
    end

    def self.finalize(engine)
      proc { self.destroy_engine(engine) }
    end

    def variant(evaluation_request)
      # TODO: create a struct for evaluation_request and marshal it to json
      self.class.variant(@engine, evaluation_request)
      # TODO: create a struct for evaluation_response and unmarshal it from json
    end

    def boolean(evaluation_request)
      # TODO: create a struct for evaluation_request and marshal it to json
      self.class.boolean(@engine, evaluation_request)
      # TODO: create a struct for evaluation_response and unmarshal it from json
    end
  end
end
