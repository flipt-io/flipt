export type Namespaces = {
  default: {
    Flags: { key: 'flag1'; value: 'variant1' | 'variant2' } | { key: 'flag2'; value: '' };
    Context: { foo: string; fizz: number; baz: boolean; };
  };
};