<template>
  <div id="main-container" class="flex flex-1">
    <AdminLteSidebar :role-mode="isKeeper ? 'keeper' : 'manager'" />

    <div id="Content" class="flex flex-col flex-1 p-6 pt-0 bg-lte-bg min-w-0">
      <MerchantTopBar
        title="Settings"
        :breadcrumbText="isKeeper ? 'Merchant Overview' : 'Overview'"
        :breadcrumbLink="isKeeper ? '/merchant-overview' : '/overview'"
      />

      <main class="flex flex-col gap-5 flex-1">
        <section class="lte-card-primary">
          <div class="lte-card-header">
            <div>
              <h2 class="lte-card-title">Application Settings</h2>
              <p class="text-xs text-lte-muted mt-0.5">
                Informasi akun aktif dan konfigurasi koneksi API Gateway.
              </p>
            </div>
          </div>

          <div class="lte-card-body space-y-5">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div class="rounded border border-lte-border p-4 bg-gray-50/50">
                <span class="text-xs text-lte-muted uppercase font-semibold">Signed-in Email</span>
                <p class="font-semibold text-base text-lte-text mt-1">{{ currentUser?.email || '-' }}</p>
              </div>
              <div class="rounded border border-lte-border p-4 bg-gray-50/50">
                <span class="text-xs text-lte-muted uppercase font-semibold">Role</span>
                <p class="font-semibold text-base text-lte-text capitalize mt-1">{{ currentUser?.roles || '-' }}</p>
              </div>
              <div class="rounded border border-lte-border p-4 bg-gray-50/50">
                <span class="text-xs text-lte-muted uppercase font-semibold">API Gateway Endpoint</span>
                <p class="font-semibold text-sm text-lte-text break-all mt-1">{{ apiBaseUrl }}</p>
              </div>
              <div class="rounded border border-lte-border p-4 bg-gray-50/50">
                <span class="text-xs text-lte-muted uppercase font-semibold">Tax Policy (POS)</span>
                <p class="font-semibold text-sm text-lte-text mt-1">PPN 11% (Server-Authoritative)</p>
              </div>
            </div>

            <div class="flex items-center gap-2 pt-2 border-t border-lte-border">
              <router-link to="/profile" class="lte-btn-primary">Manage Profile</router-link>
              <button @click="handleLogout" class="lte-btn-danger">Sign Out</button>
            </div>
          </div>
        </section>
      </main>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { APP_CONFIG } from '@/config/app'
import AdminLteSidebar from '@/components/AdminLteSidebar.vue'
import MerchantTopBar from '@/components/MerchantTopBar.vue'

const authStore = useAuthStore()
const router = useRouter()

const currentUser = computed(() => authStore.currentUser)
const isKeeper = computed(() => (authStore.userRole || '').toLowerCase() === 'keeper')
const apiBaseUrl = APP_CONFIG.API.BASE_URL

const handleLogout = async () => {
  await authStore.logout()
  router.push('/')
}
</script>
