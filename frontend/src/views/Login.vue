<template>
  <main class="min-h-screen w-full flex flex-col items-center justify-center bg-[#e9ecef] p-4">
    <!-- AdminLTE Login Box -->
    <div class="w-full max-w-[380px]">
      <!-- Brand Header -->
      <div class="flex flex-col items-center mb-5 text-center">
        <BrandLogo size="lg" />
        <p class="text-sm text-lte-muted mt-1.5">
          Warehouse &amp; Retail POS Platform
        </p>
      </div>

      <!-- Login Card -->
      <div class="bg-white rounded border border-lte-border border-t-4 border-t-lte-primary shadow-md overflow-hidden">
        <div class="p-6">
          <p class="text-center text-sm text-lte-muted mb-5">
            Sign in to start your session
          </p>

          <!-- Error Message -->
          <div
            v-if="errorMessage"
            class="mb-4 px-3.5 py-2.5 rounded border border-red-300 bg-red-50 text-red-700 text-xs font-medium"
          >
            {{ errorMessage }}
          </div>

          <form @submit.prevent="handleLogin" class="space-y-4">
            <!-- Email Input Group -->
            <div>
              <label for="login-email" class="block text-xs font-semibold text-lte-text mb-1">
                Email
              </label>
              <div class="flex items-stretch">
                <input
                  id="login-email"
                  v-model="form.email"
                  type="email"
                  class="flex-1 min-w-0 px-3 py-2 text-sm text-lte-text bg-white border border-r-0 border-[#ced4da] rounded-l focus:outline-none focus:border-lte-primary transition-colors"
                  placeholder="Email address"
                  required
                  :disabled="loading"
                >
                <span class="inline-flex items-center justify-center w-10 bg-gray-100 border border-[#ced4da] rounded-r text-lte-muted">
                  <img
                    src="@/assets/images/icons/sms-grey.svg"
                    class="size-4"
                    alt="email"
                    @error="handleIconError"
                  >
                </span>
              </div>
            </div>

            <!-- Password Input Group -->
            <div>
              <label for="login-password" class="block text-xs font-semibold text-lte-text mb-1">
                Password
              </label>
              <div class="flex items-stretch">
                <input
                  id="login-password"
                  v-model="form.password"
                  :type="showPassword ? 'text' : 'password'"
                  class="flex-1 min-w-0 px-3 py-2 text-sm text-lte-text bg-white border border-r-0 border-[#ced4da] rounded-l focus:outline-none focus:border-lte-primary transition-colors"
                  placeholder="Password"
                  required
                  :disabled="loading"
                >
                <button
                  @click="togglePassword"
                  type="button"
                  class="inline-flex items-center justify-center w-10 bg-gray-100 border border-[#ced4da] rounded-r text-lte-muted hover:bg-gray-200 transition-colors"
                  :title="showPassword ? 'Hide password' : 'Show password'"
                  :disabled="loading"
                >
                  <img src="@/assets/images/icons/eye-grey.svg" class="size-4" alt="toggle password">
                </button>
              </div>
            </div>

            <div class="flex items-center justify-between pt-1">
              <a href="#" class="text-xs text-lte-primary hover:underline">
                Forget Password?
              </a>
              <button
                type="submit"
                class="px-5 py-2 rounded bg-lte-primary hover:bg-lte-primary-hover text-white text-sm font-semibold inline-flex items-center justify-center gap-2 transition-colors disabled:opacity-60"
                :disabled="loading"
              >
                <span
                  v-if="loading"
                  class="size-4 border-2 border-white border-t-transparent rounded-full animate-spin"
                ></span>
                {{ loading ? 'Signing In...' : 'Sign In' }}
              </button>
            </div>
          </form>
        </div>
      </div>

      <p class="text-center text-xs text-lte-muted mt-4">
        &copy; CODINA Warehouse &amp; POS
      </p>
    </div>
  </main>
</template>

<script>
import { useAuthStore } from '@/stores/auth'
import BrandLogo from '@/components/BrandLogo.vue'

export default {
  name: 'Login',
  components: {
    BrandLogo
  },
  data() {
    return {
      form: {
        email: '',
        password: ''
      },
      showPassword: false,
      loading: false,
      errorMessage: ''
    }
  },
  methods: {
    async handleLogin() {
      this.loading = true
      this.errorMessage = ''

      try {
        const authStore = useAuthStore()
        await authStore.login(this.form)

        // Redirect berdasarkan role user
        const redirectUrl = authStore.getRedirectUrl()
        this.$router.push(redirectUrl)
      } catch (error) {
        console.error('Login failed:', error)

        // Handle specific error cases
        if (error.message.includes('fetch') || error.message.includes('network')) {
          this.errorMessage = 'Tidak dapat terhubung ke server. Pastikan server API berjalan.'
        } else if (error.message.includes('401') || error.message.includes('Unauthorized')) {
          this.errorMessage = 'Email atau password salah. Silakan coba lagi.'
        } else {
          this.errorMessage = error.message || 'Login gagal. Silakan coba lagi.'
        }
      } finally {
        this.loading = false
      }
    },
    togglePassword() {
      this.showPassword = !this.showPassword
    },
    handleIconError(event) {
      event.target.style.display = 'none'
    }
  }
}
</script>
