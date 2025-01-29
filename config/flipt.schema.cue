package flipt

import "strings"

import "list"

#FliptSpec: {
	// flipt-schema-v2
	//
	// Flipt config file is a YAML file defining how to configure the
	// Flipt application.
	@jsonschema(schema="http://json-schema.org/draft/2019-09/schema#")
	version:         "2.0" | *"2.0"
	experimental?:   #experimental
	analytics?:      #analytics
	authentication?: #authentication
	authorization?:  #authorization
	cors?:           #cors
	diagnostics?:    #diagnostics
	environments?:   #environments
	storage?:        #storage
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

			storage?: *{
				type:     "memory"
				cleanup?: {
					grace_period?: =~#duration | int | *"30m"
				}
			} | {
				type:     "redis"
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
				cleanup?: {
					grace_period?: =~#duration | int | *"30m"
				}
			}
		}


		methods?: {
			token?: {
				enabled?: bool | *false
				storage?: {
					type: "static"
					tokens: [
						...{
							name:       string
							credential: string
							metadata: [string]: string
						},
					]
				}
			}

			oidc?: {
				enabled?: bool | *false
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

		#authentication_oidc_provider: {
			@jsonschema(id="authentication_oidc_provider")
			issuer_url?:       string
			client_id?:        string
			client_secret?:    string
			redirect_address?: string
			nonce?:            string
			scopes?: [...string]
			use_pkce?: bool
		}

	}

	#authorization: {
		required?: bool | *false
		backend:   "local" | "bundle" | *""
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
		bundle?: {
			configuration: string
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
			"X-Flipt-Namespace",
			"X-Flipt-Accept-Server-Version",
		]
	}

	#diagnostics: {
		profiling?: {
			enabled?: bool | *true
		}
	}

	#environments: [string]: {
		default:   bool | *false
		storage:   string
		directory: string | *""
	}

	#storage: [string]: {
		remote?: string
		backend?: {
			type:  *"memory" | "local"
			path?: string
		}
		branch?:            string | *"main"
		poll_interval?:     =~#duration | *"30s"
		ca_cert_path?:      string
		ca_cert_bytes?:     string
		insecure_skip_tls?: bool | *false
		authentication?: ({
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
		publishers?: {
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
		}
	}

	_#lower: ["debug", "error", "fatal", "info", "panic", "warn"]
	_#all: list.Concat([_#lower, [for x in _#lower {strings.ToUpper(x)}]])
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
		sampling_ratio?: float & >=0 & <=1 | *1
		propagators?: [
			..."tracecontext" | "baggage" | "b3" | "b3multi" | "jaeger" | "xray" | "ottrace" | "none",
		] | *["tracecontext", "baggage"]

		otlp?: {
			endpoint?: string | *"localhost:4317"
			headers?: [string]: string
		}
	}

	#ui: {
		default_theme?: "light" | "dark" | *"system"
		topbar?: {
			color?: string
			label?: string
		}
	}

	#analytics: {
		storage?: {
			clickhouse?: {
				enabled?: bool | *false
				url?:     string | *""
			}
			prometheus?: {
				enabled?: bool | *false
				url?:     string | *""
				headers?: [string]: string
			}
		}
		buffer?: {
			capacity?:     int
			flush_period?: string | *"2m"
		}
	}

	#experimental: {}

	#duration: "^([0-9]+(ns|us|µs|ms|s|m|h))+$"
}
