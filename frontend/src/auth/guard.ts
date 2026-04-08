// ABOUTME: AuthGuard component that gates app rendering behind Auth0 authentication.
// ABOUTME: Shows a loading spinner while checking auth, redirects to login if needed.

import {defineComponent, h, watchEffect} from 'vue'
import {useAuth0} from '@auth0/auth0-vue'

export const AuthGuard = defineComponent({
  name: 'AuthGuard',
  setup(_props, {slots}) {
    const {isAuthenticated, isLoading, loginWithRedirect} = useAuth0()

    watchEffect(() => {
      if (!isLoading.value && !isAuthenticated.value) {
        loginWithRedirect()
      }
    })

    return () => {
      if (isLoading.value) {
        return h('div', {class: 'auth-loading'}, 'Loading...')
      }
      if (isAuthenticated.value && slots.default) {
        return slots.default()
      }
      return h('div', {class: 'auth-loading'}, 'Redirecting to login...')
    }
  },
})
