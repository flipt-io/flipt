import { createContext, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { getNamespace } from '~/data/api';
import { INamespace } from '~/types/Namespace';

interface NamespaceContextType {
  currentNamespace: INamespace;
}

export const NamespaceContext = createContext({
  currentNamespace: {} as INamespace
} as NamespaceContextType);

const defaultNamespace = 'default';

export default function NamespaceProvider({
  children
}: {
  children: React.ReactNode;
}) {
  let { namespaceKey } = useParams();

  const [currentNamespace, setCurrentNamespace] = useState<INamespace>(
    {} as INamespace
  );

  useEffect(() => {
    if (namespaceKey === '' || namespaceKey === undefined) {
      getNamespace(defaultNamespace).then((namespace: INamespace) => {
        setCurrentNamespace(namespace);
      });
      return;
    }

    if (namespaceKey !== undefined && namespaceKey !== currentNamespace.key) {
      getNamespace(namespaceKey).then((namespace: INamespace) => {
        setCurrentNamespace(namespace);
      });
    }
  }, [currentNamespace.key, namespaceKey]);

  return (
    <NamespaceContext.Provider
      value={{
        currentNamespace
      }}
    >
      {children}
    </NamespaceContext.Provider>
  );
}
