// ABOUTME: Auth0 Vue plugin configuration for OmniCollect.
// ABOUTME: Reads domain, clientId, and audience from Vite env vars; wires token injection.

import {createAuth0} from '@auth0/auth0-vue'
import {setTokenGetter} from '../api/client'

const domain = (import.meta as any).env?.VITE_AUTH0_DOMAIN || ''
const clientId = (import.meta as any).env?.VITE_AUTH0_CLIENT_ID || ''
const audience = (import.meta as any).env?.VITE_AUTH0_AUDIENCE || ''

export const isAuthConfigured = !!(domain && clientId)

export const auth0Plugin = isAuthConfigured
  ? createAuth0({
      domain,
      clientId,
      authorizationParams: {
        redirect_uri: window.location.origin,
        audience,
      },
      cacheLocation: 'memory',
    })
  : null

// Wire up token injection for the API client.
// The getAccessTokenSilently function is available after the plugin is installed
// and the user is authenticated. We set it up lazily on first API call.
if (auth0Plugin) {
  const client = auth0Plugin
  setTokenGetter(async () => {
    const token = await client.getAccessTokenSilently()
    if (!token) throw new Error('Authentication did not provide an access token')
    return token
  })
}
