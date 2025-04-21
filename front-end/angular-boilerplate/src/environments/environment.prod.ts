// environment.prod.ts (production)
import { env } from './.env';

export const environment = {
  production: true,
  version: env['npm_package_version'],
  defaultLanguage: 'de-DE',
  supportedLanguages: ['de-DE', 'en-US', 'es-ES', 'fr-FR', 'it-IT'],

  // Add API configuration
  apiUrl: env['API_URL'],
  apiEndpoints: {
    auth: env['API_URL'] + env['AUTH_API_PATH'],
    users: env['API_URL'] + env['USER_API_PATH'],
  },
};
