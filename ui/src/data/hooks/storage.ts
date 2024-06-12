import { Buffer } from 'buffer';
import { useState } from 'react';

const useStorage = (
  key: string,
  initialValue: any,
  storage: Storage,
  encode = false
) => {
  const [storedValue, setStoredValue] = useState(() => {
    try {
      const item = storage.getItem(key);

      if (encode) {
        const buffer = item ? Buffer.from(item, 'base64') : null;
        return buffer ? JSON.parse(buffer.toLocaleString()) : initialValue;
      }

      return item ? JSON.parse(item) : initialValue;
    } catch (error) {
      console.error(error);
      return initialValue;
    }
  });

  const setValue = (value: any) => {
    try {
      const valueToStore =
        value instanceof Function ? value(storedValue) : value;
      setStoredValue(valueToStore);

      let v = JSON.stringify(valueToStore);

      if (encode) {
        const buffer = Buffer.from(v);
        v = buffer.toString('base64');
      }

      storage.setItem(key, v);
    } catch (error) {
      console.error(error);
    }
  };

  const clearValue = () => {
    try {
      storage.removeItem(key);
    } catch (error) {
      console.error(error);
    }
    setValue(initialValue);
  };

  return [storedValue, setValue, clearValue];
};

export const useSessionStorage = (
  key: string,
  initialValue: any,
  encode = false
) => {
  return useStorage(key, initialValue, window.sessionStorage, encode);
};

export const useLocalStorage = (
  key: string,
  initialValue: any,
  encode = false
) => {
  return useStorage(key, initialValue, window.localStorage, encode);
};
