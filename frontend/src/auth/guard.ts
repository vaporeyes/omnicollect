// ABOUTME: Authentication boundary outside App setup, keyed by account-session epoch.
// ABOUTME: Disposes private stores before mounting the next account and handles login failures.
import {defineComponent, h, watch, ref} from 'vue'
import {getActivePinia} from 'pinia'
import {useAuth0} from '@auth0/auth0-vue'
import {sessionVersion, sessionReady, logoutRequested, sessionFailure, setSessionIdentity} from './session'
import {useCollectionStore} from '../stores/collectionStore'
import {useModuleStore} from '../stores/moduleStore'
import {useSelectionStore} from '../stores/selectionStore'
import {useSmartFolderStore} from '../stores/smartFolderStore'
import {useToastStore} from '../stores/toastStore'
import {useSummaryStore} from '../stores/summaryStore'
export const AuthGuard = defineComponent({
  name: 'AuthGuard',
  setup(_props, {slots}) {
    const auth = useAuth0()
    const pinia = getActivePinia()!
    const redirecting = ref(false)
    async function login() {
      if (redirecting.value) return
      redirecting.value = true
      sessionFailure.value = ''
      try {await auth.loginWithRedirect()}
      catch {sessionFailure.value = 'Sign in could not start. Check your connection and retry.'}
      finally {redirecting.value = false}
    }
    watch(sessionVersion, () => {
      // List every account-scoped store here; disposal must remove registry entries before remount.
      const stores = [useCollectionStore(pinia), useModuleStore(pinia), useSelectionStore(pinia), useSmartFolderStore(pinia), useToastStore(pinia), useSummaryStore(pinia)]
      useSelectionStore(pinia).clear()
      for (const store of stores) store.$dispose()
      pinia.state.value = {}
    }, {flush: 'sync'})
    watch(() => [auth.isLoading.value, auth.isAuthenticated.value, auth.user.value?.sub, logoutRequested.value] as const, ([loading, authenticated, subject, signingOut]) => {
      setSessionIdentity(!loading && authenticated && subject && !signingOut ? `auth:${subject}` : null)
      if (!loading && !authenticated && !signingOut) void login()
    }, {immediate: true, flush: 'sync'})
    return () => {
      if (sessionReady.value && auth.isAuthenticated.value && !auth.isLoading.value) {
        return h('div', {key: sessionVersion.value}, slots.default?.())
      }
      return h('div', {class: 'auth-loading', role: 'status'}, [
        h('p', sessionFailure.value || (auth.isLoading.value ? 'Checking account…' : logoutRequested.value ? 'Signed out or signing out. Reload to check the session.' : 'Waiting for sign in…')),
        !auth.isLoading.value && !redirecting.value ? h('button', {onClick: () => {logoutRequested.value = false; void login()}}, 'Sign in') : null,
        h('button', {onClick: () => window.location.reload()}, 'Reload'),
      ])
    }
  },
})
