export interface IAuthMethod {
  method: 'METHOD_TOKEN' | 'METHOD_OIDC';
  enabled: boolean;
  sessionCompatible: boolean;
  metadata: { [key: string]: any };
}

export interface IAuthMethodList {
  methods: IAuthMethod[];
}

export interface IAuth {
  id: string;
  method: string;
  expiresAt: string;
  createdAt: string;
  updatedAt: string;
}
