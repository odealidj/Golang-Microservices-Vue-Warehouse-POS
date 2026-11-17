<template>
  <!-- AdminLTE Top Navbar -->
  <header class="sticky top-0 z-30 flex items-center justify-between h-[57px] bg-white border-b border-lte-border px-4 lg:px-6 -mx-6 mb-5">
    <!-- Left Navbar Links -->
    <div class="flex items-center gap-3">
      <button
        type="button"
        @click="toggleSidebar"
        class="inline-flex items-center justify-center size-9 rounded text-lte-muted hover:text-lte-text hover:bg-gray-100 transition-colors"
        title="Toggle Sidebar"
        aria-label="Toggle Sidebar"
      >
        <svg class="size-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
        </svg>
      </button>

      <nav class="hidden sm:flex items-center gap-4 text-sm text-lte-muted">
        <router-link :to="homeRoute" class="hover:text-lte-text transition-colors font-medium">
          Home
        </router-link>
        <router-link
          v-if="isKeeper"
          to="/transactions/add-1"
          class="hover:text-lte-primary transition-colors font-medium"
        >
          POS Checkout
        </router-link>
        <router-link to="/settings" class="hover:text-lte-text transition-colors font-medium">
          Settings
        </router-link>
      </nav>
    </div>

    <!-- Right Navbar Links -->
    <div class="flex items-center gap-3">
      <router-link
        to="/profile"
        class="flex items-center gap-2.5 py-1 px-2 rounded hover:bg-gray-100 transition-colors"
      >
        <div v-if="userLoading" class="size-8 rounded-full bg-gray-200 animate-pulse"></div>
        <img
          v-else
          :src="currentUser?.avatar || defaultAvatar"
          class="size-8 rounded-full object-cover border border-lte-border"
          alt="User photo"
        >
        <div class="hidden md:flex flex-col leading-tight text-left">
          <span class="text-sm font-medium text-lte-text max-w-[150px] truncate">
            {{ currentUser?.name || userName }}
          </span>
          <span class="text-[11px] text-lte-muted capitalize">
            {{ currentUser?.roles || userRole }}
          </span>
        </div>
      </router-link>

      <button
        @click="handleLogout"
        class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded text-xs font-semibold text-red-600 border border-red-200 bg-red-50 hover:bg-red-600 hover:text-white transition-colors cursor-pointer"
        title="Logout"
        type="button"
      >
        <svg class="size-4 shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
        </svg>
        <span class="hidden sm:inline">Logout</span>
      </button>
    </div>
  </header>

  <!-- AdminLTE Content Header (Title + Breadcrumb) -->
  <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 mb-5">
    <h1 class="text-2xl font-medium text-lte-text tracking-tight">
      {{ title }}
    </h1>

    <ol class="flex items-center gap-1.5 text-sm text-lte-muted">
      <li>
        <router-link :to="homeRoute" class="text-lte-primary hover:underline">
          Home
        </router-link>
      </li>
      <template v-if="breadcrumbText">
        <li class="text-gray-400">/</li>
        <li>
          <router-link :to="breadcrumbLink" class="text-lte-primary hover:underline">
            {{ breadcrumbText }}
          </router-link>
        </li>
      </template>
      <li class="text-gray-400">/</li>
      <li class="text-lte-muted font-normal truncate max-w-[200px]">
        {{ title }}
      </li>
    </ol>
  </div>

  <!-- Logout Confirmation Modal -->
  <div v-if="showLogoutConfirm" class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <div @click="cancelLogout" class="absolute inset-0 bg-black/50"></div>
    <div class="relative bg-white rounded border border-lte-border shadow-xl w-full max-w-sm overflow-hidden">
      <div class="flex items-center justify-between px-5 py-3.5 border-b border-lte-border bg-gray-50">
        <h3 class="font-semibold text-base text-lte-text">Konfirmasi Logout</h3>
        <button @click="cancelLogout" type="button" class="text-gray-400 hover:text-gray-600 text-lg leading-none">
          &times;
        </button>
      </div>
      <div class="p-5 text-sm text-lte-muted">
        Apakah Anda yakin ingin keluar dari sesi aplikasi saat ini?
      </div>
      <div class="flex justify-end gap-2 px-5 py-3 bg-gray-50 border-t border-lte-border">
        <button
          @click="cancelLogout"
          type="button"
          class="px-4 py-2 rounded border border-lte-border bg-white text-sm font-medium text-lte-text hover:bg-gray-100 transition-colors"
        >
          Batal
        </button>
        <button
          @click="confirmLogout"
          type="button"
          class="px-4 py-2 rounded bg-lte-danger text-white text-sm font-medium hover:bg-red-700 transition-colors"
        >
          Logout
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { removeAuthToken } from '@/js/api'
import { toggleSidebar } from '@/js/layoutState'
import defaultAvatar from '@/assets/images/photos/photos-1.png'

defineProps({
  title: {
    type: String,
    default: 'Dashboard'
  },
  userName: {
    type: String,
    default: 'User'
  },
  userRole: {
    type: String,
    default: 'User'
  },
  breadcrumbText: {
    type: String,
    default: null
  },
  breadcrumbLink: {
    type: String,
    default: '/overview'
  }
})

const router = useRouter()
const authStore = useAuthStore()

const userLoading = ref(false)
const showLogoutConfirm = ref(false)

const currentUser = computed(() => authStore.currentUser)
const isKeeper = computed(() => (authStore.userRole || '').toLowerCase() === 'keeper')
const homeRoute = computed(() => (isKeeper.value ? '/merchant-overview' : '/overview'))

const handleLogout = () => {
  showLogoutConfirm.value = true
}

const confirmLogout = async () => {
  try {
    await authStore.logout()
    removeAuthToken()
    router.push('/')
  } catch (error) {
    console.error('Error during logout:', error)
    router.push('/')
  } finally {
    showLogoutConfirm.value = false
  }
}

const cancelLogout = () => {
  showLogoutConfirm.value = false
}

const handleKeydown = (event) => {
  if (event.key === 'Escape' && showLogoutConfirm.value) {
    cancelLogout()
  }
}

onMounted(async () => {
  if (authStore.isLoggedIn && !authStore.currentUser) {
    userLoading.value = true
    try {
      await authStore.fetchCurrentUser()
    } catch (error) {
      console.error('Error loading user data:', error)
    } finally {
      userLoading.value = false
    }
  }
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>
