import '@testing-library/jest-dom';

// Polyfill TextEncoder/TextDecoder for jest-environment-jsdom + react-router
import { TextEncoder, TextDecoder } from 'util';
Object.assign(global, { TextEncoder, TextDecoder });
