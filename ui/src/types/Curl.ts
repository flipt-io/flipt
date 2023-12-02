export interface ICurlOptions {
  method: 'GET' | 'POST'; // maybe we'll need to extend this in the future
  headers?: Record<string, string>;
  body?: any;
  uri: string;
}
