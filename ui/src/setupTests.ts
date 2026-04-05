// eslint-disable-next-line import/no-extraneous-dependencies
import '@testing-library/jest-dom';
// Polyfill TextEncoder/TextDecoder for jest-environment-jsdom + react-router
import { TextDecoder, TextEncoder } from 'util';

Object.assign(global, { TextEncoder, TextDecoder });
