// environment.ts (development)
import { env } from './.env';

export const environment = {
  production: false,
  version: env['npm_package_version'] + '-dev',
  defaultLanguage: 'de-DE',
  supportedLanguages: ['de-DE', 'en-US', 'es-ES', 'fr-FR', 'it-IT'],
  disabledLogChannels: [], // ['ProductListComponent', 'SomeOtherChannel']

  // Add API configuration
  apiUrl: env['API_URL'],
  apiEndpoints: {
    users: env['API_URL'] + env['USER_API_PATH'],
    orders: env['API_URL'] + env['ORDER_API_PATH'],
    inventories: env['API_URL'] + env['INVENTORY_API_PATH'],
  },

  // WebSocket configuration
  websocketUrl: env['WEBSOCKET_URL'],
};
