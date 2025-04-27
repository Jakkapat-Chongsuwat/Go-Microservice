// src/environments/.env.ts

// This file exports environment variables for use in environment configurations

// Read package version from package.json
import { version } from '../../package.json';

export const env: { [key: string]: string } = {
  'npm_package_version': version,

  'API_URL': 'http://localhost:45575',
  'USER_API_PATH': '/users',
  'ORDER_API_PATH': '/orders',
  'INVENTORY_API_PATH': '/products',

  'WEBSOCKET_URL': 'ws://ws.192.168.49.2.nip.io/ws',
};
