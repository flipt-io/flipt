import { Session } from '~/types/auth/Session';
import { User } from '~/types/auth/User';

export function getUser(session?: Session): User | undefined {
  if (session?.self) {
    const self = session.self;
    const u = {} as User;
    const authMethods = ['github', 'oidc', 'jwt'];
    const authMethod = authMethods.find(
      (method) => `METHOD_${method.toLocaleUpperCase()}` === self.method
    );

    if (authMethod) {
      u.authMethod = authMethod;
      const metadata = self.metadata;

      const authMethodNameKey = `io.flipt.auth.${authMethod}.name`;
      u.name = metadata[authMethodNameKey as keyof typeof metadata] ?? 'User';

      const authMethodPictureKey = `io.flipt.auth.${authMethod}.picture`;
      if (metadata[authMethodPictureKey as keyof typeof metadata]) {
        u.imgURL = metadata[authMethodPictureKey as keyof typeof metadata];
      }

      const authMethodPreferredUsernameKey = `io.flipt.auth.${authMethod}.preferred_username`;
      if (metadata[authMethodPreferredUsernameKey as keyof typeof metadata]) {
        u.login =
          metadata[authMethodPreferredUsernameKey as keyof typeof metadata];
      }
      const authMethodIssuerKey = `io.flipt.auth.${authMethod}.issuer`;
      if (metadata[authMethodIssuerKey as keyof typeof metadata]) {
        u.issuer = metadata[authMethodIssuerKey as keyof typeof metadata];
      }
    }
    return u;
  }
  return undefined;
}
