package flipt

#FliptSpec: {
	// Flipt Configuration
	//
	// Flipt config file is a YAML file defining how to configure the
	// Flipt application.
	@jsonschema(schema="http://json-schema.org/draft/2019-09/schema#")
	version?: string | *"1.0"
	authentication?: #authentication
	cache?:          #cache
	cors?:           #cors
	db?:             #db
	log?:            #log
	meta?:           #meta
	server?:         #server
	tracing?:        #tracing
	ui?:             #ui

	#authentication: {
		required?: bool | *false

		// Methods
		methods?: {
			// Token
			token?: {
				enabled?: bool | *false
				cleanup?: #authentication.#authentication_cleanup
			}
		}

		#authentication_cleanup: {
			@jsonschema(id="authentication_cleanup")
			interval?:     #duration | *"1h"
			grace_period?: #duration | *"30m"
		}
	}

	#cache: {
		enabled?: bool | *false
		backend?: "memory" | "redis" | *"memory"
		ttl?:     #duration | *"60s"

		// Redis
		redis?: {
			host?:     string | *"localhost"
			port?:     int | *6379
			db?:       int | *0
			password?: string
		}

		// Memory
		memory?: {
			eviction_interval?: #duration | *"5m"
		}
	}

	#cors: {
		enabled?:         bool | *false
		allowed_origins?: [...] | *["*"]
	}

	#db: {
		url?:               string | *"file:/var/opt/flipt/flipt.db"
		protocol?:          "cockroach" | "cockroachdb" | "file" | "mysql" | "postgres" | "sqlite"
		host?:              string
		port?:              int
		name?:              string
		user?:              string
		password?:          string
		max_idle_conn?:     int | *2
		max_open_conn?:     int
		conn_max_lifetime?: int
	}

	#log: {
		file?:       string
		encoding?:   "json" | "console" | *"console"
		level?:      #log.#log_level
		grpc_level?: #log.#log_level

		#log_level: "debug" | "DEBUG" | "error" | "ERROR" | "fatal" | "FATAL" | "info" | "INFO" | "PANIC" | "panic" | "trace" | "TRACE" | "warn" | "WARN"
	}

	#meta: {
		check_for_updates?: bool | *true
		telemetry_enabled?: bool | *true
		state_directory?:   string | *"$HOME/.config/flipt"
	}

	#server: {
		protocol?:   "http" | "https" | *"http"
		host?:       string | *"0.0.0.0"
		https_port?: int | *443
		http_port?:  int | *8080
		grpc_port?:  int | *9000
		cert_file?:  string
		cert_key?:   string
	}

	#tracing: {
		// Jaeger
		jaeger?: {
			enabled?: bool | *false
			host?:    string | *"localhost"
			port?:    int | *6831
		}
	}

	#ui: enabled?: bool | *true

	#duration: =~"^([0-9]+(ns|us|µs|ms|s|m|h))+$" | int
}
