<template>
  <div id="main-container" class="flex flex-1">
    <MerchantSidebar v-if="isKeeper" />
    <aside v-else class="w-[280px] shrink-0 bg-white border-r border-monday-border p-6 flex flex-col gap-6">
      <img src="@/assets/images/logos/logo.svg" class="h-8 w-fit" alt="logo">
      <router-link to="/overview" class="btn btn-primary font-semibold text-center">
        Back to Overview
      </router-link>
    </aside>

    <div id="Content" class="flex flex-col flex-1 p-6 pt-0">
      <MerchantTopBar
        title="Settings"
        :breadcrumbText="isKeeper ? 'Back to Merchant Overview' : 'Back to Overview'"
        :breadcrumbLink="isKeeper ? '/merchant-overview' : '/overview'"
      />

      <main class="flex flex-col gap-6 flex-1">
        <section class="bg-white rounded-3xl p-6 flex flex-col gap-6">
          <div>
            <h2 class="font-bold text-xl">Application Settings</h2>
            <p class="text-monday-gray text-sm mt-1">
              Informasi akun aktif dan konfigurasi koneksi API Gateway.
            </p>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="rounded-2xl border border-monday-border p-4 flex flex-col gap-1">
              <span class="text-sm text-monday-gray">Signed-in Email</span>
              <span class="font-semibold text-lg">{{ currentUser?.email || '-' }}</span>
            </div>
            <div class="rounded-2xl border border-monday-border p-4 flex flex-col gap-1">
              <span class="text-sm text-monday-gray">Role</span>
              <span class="font-semibold text-lg capitalize">{{ currentUser?.roles || '-' }}</span>
            </div>
            <div class="rounded-2xl border border-monday-border p-4 flex flex-col gap-1">
              <span class="text-sm text-monday-gray">API Gateway Endpoint</span>
              <span class="font-semibold text-base break-all">{{ apiBaseUrl }}</span>
            </div>
            <div class="rounded-2xl border border-monday-border p-4 flex flex-col gap-1">
              <span class="text-sm text-monday-gray">Tax Policy (POS)</span>
              <span class="font-semibold text-base">PPN 11% (Server-Authoritative)</span>
            </div>
          </div>

          <div class="flex items-center gap-3">
            <router-link to="/profile" class="btn btn-primary font-semibold">
              Manage Profile
            </router-link>
            <button @click="handleLogout" class="btn btn-red font-semibold">
              Sign Out
            </button>
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
import MerchantSidebar from '@/components/MerchantSidebar.vue'
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
