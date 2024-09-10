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
	analytics:       #analytics
	audit?:          #audit
	authentication?: #authentication
	authorization?:  #authorization
	cache?:          #cache
	cloud?:          #cloud
	cors?:           #cors
	diagnostics?:    #diagnostics
	storage?:        #storage
	db?:             #db
	log?:            #log
	meta?:           #meta
	server?:         #server
	metrics?:        #metrics
	tracing?:        #tracing
	ui?:             #ui

	#authentication: {
		required?: bool | *false
		exclude?: {
			management: bool | *false
			metadata:   bool | *false
			evaluation: bool | *false
			ofrep:      bool | *false
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
					metadata?: [string]: string
				}
			}

			cloud?: {
				enabled?: bool | *false
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

			github?: {
				enabled?:          bool | *false
				server_url?:       string
				api_url?:          string
				client_secret?:    string
				client_id?:        string
				redirect_address?: string
				scopes?: [...string]
				allowed_organizations?: [...] | string
				allowed_teams?: [string]: [...string]
			}

			jwt?: {
				enabled?: bool | *false
				validate_claims?: {
					issuer?:  string
					subject?: string
					audiences?: [...string]
				}
				jwks_url?:        string
				public_key_file?: string
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
			nonce?: string
			scopes?: [...string]
			use_pkce?: bool
		}
	}

	#authorization: {
		required?: bool | *false
		backend:   "local" | "object" | "bundle" | "cloud" | *""
		local?: {
			policy?: {
				poll_interval: =~#duration | *"5m"
				path:          string
			}
			data?: {
				poll_interval: =~#duration | *"5m"
				path:          string
			}
		}
		object?: {
			type: "s3" | *""
			s3?: {
				region:    string
				bucket:    string
				prefix?:   string
				endpoint?: string
			}
		}
		bundle?: {
			configuration: string
		}
		cloud?: {
			poll_interval: =~#duration | *"5m"
		}
	}

	#cache: {
		enabled?: bool | *false
		backend?: *"memory" | "redis"
		ttl?:     =~#duration | int | *"60s"

		redis?: {
			host?:               string | *"localhost"
			port?:               int | *6379
			require_tls?:        bool | *false
			db?:                 int | *0
			username?:           string
			password?:           string
			pool_size?:          int | *0
			min_idle_conn?:      int | *0
			conn_max_idle_time?: =~#duration | int | *0
			net_timeout?:        =~#duration | int | *0
			ca_cert_path?:       string
			ca_cert_bytes?:      string
			insecure_skip_tls?:  bool | *false
		}

		memory?: {
			enabled?:           bool | *false
			eviction_interval?: =~#duration | int | *"5m"
			expiration?:        =~#duration | int | *"60s"
		}
	}

	#cloud: {
		host?:         string | *"flipt.cloud"
		organization?: string
		gateway?:      string
		authentication?: {
			api_key?: string
		}
	}

	#cors: {
		enabled?: bool | *false
		allowed_origins?: [...] | string | *["*"]
		allowed_headers?: [...string] | string | *[
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Fern-Language",
			"X-Fern-SDK-Name",
			"X-Fern-SDK-Version",
			"X-Flipt-Namespace",
			"X-Flipt-Accept-Server-Version",
		]
	}

	#diagnostics: {
		profiling?: {
			enabled?: bool | *true
		}
	}

	#storage: {
		type:       "database" | "git" | "local" | "object" | "oci" | *""
		read_only?: bool | *false
		local?: path: string | *"."
		git?: {
			repository:         string
			backend?:           *"memory" | "local"
			ref?:               string | *"main"
			ref_type?:          *"static" | "semver"
			directory?:         string
			poll_interval?:     =~#duration | *"30s"
			ca_cert_path?:      string
			ca_cert_bytes?:     string
			insecure_skip_tls?: bool | *false
			authentication?:    ({
				basic: {
					username: string
					password: string
				}
			} | {
				token: access_token: string
			} | {
				ssh: {
					user?:            string | *"git"
					password:         string
					private_key_path: string
				}
			} | {
				ssh: {
					user?:             string | *"git"
					password:          string
					private_key_bytes: string
				}
			})
		}
		object?: {
			type: "s3" | "azblob" | "googlecloud" | *""
			s3?: {
				region:         string
				bucket:         string
				prefix?:        string
				endpoint?:      string
				poll_interval?: =~#duration | *"1m"
			}
			azblob?: {
				container:      string
				endpoint?:      string
				poll_interval?: =~#duration | *"1m"
			}
			googlecloud?: {
				bucket:         string
				prefix?:        string
				poll_interval?: =~#duration | *"1m"
			}
		}
		oci?: {
			repository:         string
			bundles_directory?: string
			authentication?: {
				type:     "aws-ecr" | *"static"
				username: string
				password: string
			}
			poll_interval?:    =~#duration | *"30s"
			manifest_version?: "1.0" | *"1.1"
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

	_#lower: ["debug", "error", "fatal", "info", "panic", "warn"]
	_#all: _#lower + [for x in _#lower {strings.ToUpper(x)}]
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
		protocol?:                *"http" | "https"
		host?:                    string | *"0.0.0.0"
		https_port?:              int | *443
		http_port?:               int | *8080
		grpc_port?:               int | *9000
		cert_file?:               string
		cert_key?:                string
		grpc_conn_max_idle_time?: =~#duration
		grpc_conn_max_age?:       =~#duration
		grpc_conn_max_age_grace?: =~#duration
		cloud?: {
			enabled?: bool | *false
			port?:    int | *8443
		}
	}

	#metrics: {
		enabled?:  bool | *true
		exporter?: *"prometheus" | "otlp"

		otlp?: {
			endpoint?: string | *"localhost:4317"
			headers?: [string]: string
		}
	}

	#tracing: {
		enabled?:        bool | *false
		exporter?:       *"jaeger" | "zipkin" | "otlp"
		sampling_ratio?: float & >=0 & <=1 | *1
		propagators?: [
			..."tracecontext" | "baggage" | "b3" | "b3multi" | "jaeger" | "xray" | "ottrace" | "none",
		] | *["tracecontext", "baggage"]

		jaeger?: {
			enabled?: bool | *false
			host?:    string | *"localhost"
			port?:    int | *6831
		}

		zipkin?: {
			endpoint?: string | *"http://localhost:9411/api/v2/spans"
		}

		otlp?: {
			endpoint?: string | *"localhost:4317"
			headers?: [string]: string
		}
	}

	#ui: {
		enabled?:       bool | *true
		default_theme?: "light" | "dark" | *"system"
		topbar?: {
			color?: string
			label?: string
		}
	}

	#audit: {
		sinks?: {
			log?: {
				enabled?:  bool | *false
				file?:     string | *""
				encoding?: *"" | "json" | "console"
			}
			webhook?: {
				enabled?:              bool | *false
				url?:                  string | *""
				max_backoff_duration?: =~#duration | *"15s"
				signing_secret?:       string | *""
				templates?: [...{
					url:  string
					body: string
					headers?: [string]: string
				}]
			}
			cloud?: {
				enabled?: bool | *false
			}
			kafka?: {
				enabled?: bool | *false
				topic:    string
				bootstrap_servers: [...string]
				encoding?: *"protobuf" | "avro"
				schema_registry?: {
					url: string
				} | null
				require_tls?:        bool | *false
				insecure_skip_tls?:  bool | *false
				authentication?: {
					username: string
					password: string
				} | null
			}
		}
		buffer?: {
			capacity?:     int | *2
			flush_period?: string | *"2m"
		}
		events?: [...string] | *["*:*"]
	}

	#analytics: {
		storage?: {
			clickhouse?: {
				enabled?: bool | *false
				url?:     string | *""
			}
		}
		buffer?: {
			capacity?:     int
			flush_period?: string | *"2m"
		}
	}

	#experimental: {
		cloud?: {
			enabled?: bool | *false
		}
	}

	#duration: "^([0-9]+(ns|us|µs|ms|s|m|h))+$"
}
