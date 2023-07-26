package flipt

import "strings"

#FliptSpec: {
	// flipt-schema-v1
	//
	// Flipt config file is a YAML file defining how to configure the
	// Flipt application.
	@jsonschema(schema="http://json-schema.org/draft/2019-09/schema#")
	version?:        "1.0" | *"1.0"
	experimental?:   #experimental
	audit?:          #audit
	authentication?: #authentication
	cache?:          #cache
	cors?:           #cors
	storage?:        #storage
	db?:             #db
	log?:            #log
	meta?:           #meta
	server?:         #server
	tracing?:        #tracing
	ui?:             #ui

	#experimental: filesystem_storage?: enabled?: bool

	#authentication: {
		required?: bool | *false
		exclude?: {
			management: bool | *false
			metadata:   bool | *false
			evaluation: bool | *false
		}
		session?: {
			domain?:        string
			secure?:        bool
			token_lifetime: =~#duration | *"24h"
			state_lifetime: =~#duration | *"10m"
			csrf?: {
				key: string
			}
		}

		methods?: {
			token?: {
				enabled?: bool | *false
				cleanup?: #authentication.#authentication_cleanup
				bootstrap?: {
					token?:     string
					expiration: =~#duration | int
				}
			}

			oidc?: {
				enabled?: bool | *false
				cleanup?: #authentication.#authentication_cleanup
				providers?: {
					{[=~"^.*$" & !~"^()$"]: #authentication.#authentication_oidc_provider}
				}
				email_matches?: [...] | string
			}

			kubernetes?: {
				enabled?:                   bool | *false
				discovery_url:              string
				ca_path:                    string
				service_account_token_path: string
				cleanup?:                   #authentication.#authentication_cleanup
			}
		}

		#authentication_cleanup: {
			@jsonschema(id="authentication_cleanup")
			interval?:     =~#duration | int | *"1h"
			grace_period?: =~#duration | int | *"30m"
		}

		#authentication_oidc_provider: {
			@jsonschema(id="authentication_oidc_provider")
			issuer_url?:       string
			client_id?:        string
			client_secret?:    string
			redirect_address?: string
		}
	}

	#cache: {
		enabled?: bool | *false
		backend?: *"memory" | "redis"
		ttl?:     =~#duration | int | *"60s"

		// Redis
		redis?: {
			host?:     string | *"localhost"
			port?:     int | *6379
			db?:       int | *0
			password?: string
		}

		// Memory
		memory?: {
			enabled?:           bool | *false
			eviction_interval?: =~#duration | int | *"5m"
			expiration?:        =~#duration | int | *"60s"
		}
	}

	#cors: {
		enabled?:         bool | *false
		allowed_origins?: [...] | string | *["*"]
	}

	#storage: {
		type: "database" | "git" | "local" | *""
		local?: path: string | *"."
		git?: {
			repository:      string
			ref?:            string | *"main"
			poll_interval?:  =~#duration | *"30s"
			authentication?: ({
				basic: {
					username: string
					password: string
				}
			} | {
				token: access_token: string
			})
		}
	}

	#db: {
		password?:                    string
		max_idle_conn?:               int | *2
		max_open_conn?:               int
		conn_max_lifetime?:           =~#duration | int
		prepared_statements_enabled?: bool | *true
	} & ({
		url?: string | *"file:/var/opt/flipt/flipt.db"
	} | {
		protocol?: *"sqlite" | "cockroach" | "cockroachdb" | "file" | "mysql" | "postgres"
		host?:     string
		port?:     int
		name?:     string
		user?:     string
	})

	_#lower: ["debug", "error", "fatal", "info", "panic", "trace", "warn"]
	_#all: _#lower + [ for x in _#lower {strings.ToUpper(x)}]
	#log: {
		file?:       string
		encoding?:   *"console" | "json"
		level?:      #log.#log_level
		grpc_level?: #log.#log_level
		keys?: {
			time?:    string | *"T"
			level?:   string | *"L"
			message?: string | *"M"
		}

		#log_level: or(_#all)
	}

	#meta: {
		check_for_updates?: bool | *true
		telemetry_enabled?: bool | *true
		state_directory?:   string | *"$HOME/.config/flipt"
	}

	#server: {
		protocol?:   *"http" | "https"
		host?:       string | *"0.0.0.0"
		https_port?: int | *443
		http_port?:  int | *8080
		grpc_port?:  int | *9000
		cert_file?:  string
		cert_key?:   string
	}

	#tracing: {
		enabled?:  bool | *false
		exporter?: *"jaeger" | "zipkin" | "otlp"

		// Jaeger
		jaeger?: {
			enabled?: bool | *false
			host?:    string | *"localhost"
			port?:    int | *6831
		}

		// Zipkin
		zipkin?: {
			endpoint?: string | *"http://localhost:9411/api/v2/spans"
		}

		// OTLP
		otlp?: {
			endpoint?: string | *"localhost:4317"
		}
	}

	#ui: enabled?: bool | *true

	#audit: {
		sinks?: {
			log?: {
				enabled?: bool | *false
				file?:    string | *""
			}
		}
		buffer?: {
			capacity?:     int | *2
			flush_period?: string | *"2m"
		}
	}

	#duration: "^([0-9]+(ns|us|µs|ms|s|m|h))+$"
}
