export function classNames(...classes: string[]) {
  return classes.filter(Boolean).join(' ');
}

export function stringAsKey(str: string) {
  return str.toLowerCase().split(/\s+/).join('-');

  // // Auto generated keys  should not begin or end in a hyphen
  // if (temp.charAt(0) == "-") {
  //   temp = temp.slice(1);
  // }

  // if (temp.charAt(temp.length - 1) == "-") {
  //   temp = temp.slice(0, -1);
  // }

  // return temp;
}

export function truncateKey(str: string, len: number = 25): string {
  return str.length > len ? str.substring(0, len) + '...' : str;
}

const namespaces = '/namespaces/';
export function addNamespaceToPath(path: string, key: string): string {
  if (path.startsWith(namespaces)) {
    // [0] before slash ('')
    // [1] /namespaces/
    // [2] namespace key
    // [...] after slash
    const [, , existingKey, ...parts] = path.split('/');
    if (existingKey === key) {
      return path;
    }
    return `${namespaces}${key}/${parts.join('/')}`;
  }
  return `${namespaces}${key}${path}`;
}
