<template>
  <div>
    <!-- Mobile backdrop -->
    <div
      v-if="mobileSidebarOpen"
      @click="closeMobileSidebar"
      class="fixed inset-0 z-40 bg-black/50 lg:hidden"
    ></div>

    <!-- AdminLTE Main Sidebar -->
    <aside
      :class="[
        'fixed lg:static inset-y-0 left-0 z-50 flex flex-col h-screen shrink-0 bg-lte-sidebar text-lte-sidebar-text shadow-xl transition-all duration-200 select-none',
        sidebarCollapsed ? 'lg:w-[72px]' : 'w-[250px]',
        mobileSidebarOpen ? 'translate-x-0 w-[250px]' : '-translate-x-full lg:translate-x-0'
      ]"
    >
      <!-- Brand Link -->
      <router-link
        :to="homeRoute"
        @click="closeMobileSidebar"
        :class="[
          'flex items-center h-[57px] shrink-0 border-b border-lte-sidebar-border px-4 transition-colors hover:bg-white/5',
          sidebarCollapsed ? 'lg:justify-center lg:px-2' : 'justify-between'
        ]"
      >
        <BrandLogo variant="light" :compact="sidebarCollapsed && !mobileSidebarOpen" />
        <span
          v-if="!sidebarCollapsed || mobileSidebarOpen"
          class="text-[10px] font-semibold uppercase tracking-wider px-1.5 py-0.5 rounded bg-lte-primary/20 text-blue-300 border border-blue-400/30"
        >
          {{ roleMode === 'keeper' ? 'POS' : 'WMS' }}
        </span>
      </router-link>

      <!-- User Panel -->
      <div
        :class="[
          'flex items-center gap-3 py-3.5 px-4 border-b border-lte-sidebar-border shrink-0',
          sidebarCollapsed ? 'lg:justify-center lg:px-2' : ''
        ]"
      >
        <img
          :src="currentUser?.avatar || defaultAvatar"
          class="size-9 rounded-full object-cover ring-2 ring-white/20 shrink-0"
          alt="User Avatar"
        >
        <div v-if="!sidebarCollapsed || mobileSidebarOpen" class="min-w-0 flex-1">
          <p class="text-sm font-medium text-white truncate">
            {{ currentUser?.name || (roleMode === 'keeper' ? 'Keeper' : 'Manager') }}
          </p>
          <p class="flex items-center gap-1.5 text-xs text-lte-sidebar-text/80 capitalize">
            <span class="inline-block size-2 rounded-full bg-lte-success"></span>
            {{ currentUser?.roles || roleMode }}
          </p>
        </div>
      </div>

      <!-- Navigation -->
      <div class="flex-1 overflow-y-auto px-2 py-3 space-y-5 lte-sidebar-scroll">
        <!-- Main Menu -->
        <div>
          <p
            v-if="!sidebarCollapsed || mobileSidebarOpen"
            class="px-3 mb-1.5 text-[11px] font-semibold uppercase tracking-wider text-lte-sidebar-text/60"
          >
            Main Menu
          </p>
          <ul class="space-y-1">
            <li v-for="item in mainMenuItems" :key="item.to">
              <router-link
                :to="item.to"
                @click="closeMobileSidebar"
                :title="sidebarCollapsed ? item.label : ''"
                :class="[
                  'flex items-center gap-3 px-3 py-2.5 rounded text-sm font-medium transition-colors',
                  isItemActive(item)
                    ? 'bg-lte-primary text-white shadow-sm'
                    : 'text-lte-sidebar-text hover:bg-lte-sidebar-hover hover:text-white',
                  sidebarCollapsed ? 'lg:justify-center lg:px-2' : ''
                ]"
              >
                <img
                  :src="item.icon"
                  class="size-[18px] shrink-0 brightness-0 invert"
                  :class="isItemActive(item) ? 'opacity-100' : 'opacity-80'"
                  alt=""
                >
                <span v-if="!sidebarCollapsed || mobileSidebarOpen" class="truncate">
                  {{ item.label }}
                </span>
              </router-link>
            </li>
          </ul>
        </div>

        <!-- Account Settings -->
        <div>
          <p
            v-if="!sidebarCollapsed || mobileSidebarOpen"
            class="px-3 mb-1.5 text-[11px] font-semibold uppercase tracking-wider text-lte-sidebar-text/60"
          >
            Account Settings
          </p>
          <ul class="space-y-1">
            <!-- Roles (Manager only) -->
            <li v-if="roleMode === 'manager'">
              <router-link
                to="/roles"
                @click="closeMobileSidebar"
                :title="sidebarCollapsed ? 'Roles' : ''"
                :class="[
                  'flex items-center gap-3 px-3 py-2.5 rounded text-sm font-medium transition-colors',
                  route.path.startsWith('/roles')
                    ? 'bg-lte-primary text-white shadow-sm'
                    : 'text-lte-sidebar-text hover:bg-lte-sidebar-hover hover:text-white',
                  sidebarCollapsed ? 'lg:justify-center lg:px-2' : ''
                ]"
              >
                <img
                  :src="icons.stickynote"
                  class="size-[18px] shrink-0 brightness-0 invert"
                  :class="route.path.startsWith('/roles') ? 'opacity-100' : 'opacity-80'"
                  alt=""
                >
                <span v-if="!sidebarCollapsed || mobileSidebarOpen" class="truncate">Roles</span>
              </router-link>
            </li>

            <!-- Manage User Treeview (Manager only) -->
            <li v-if="roleMode === 'manager'">
              <button
                type="button"
                @click="handleUserMenuClick"
                :title="sidebarCollapsed ? 'Manage User' : ''"
                :class="[
                  'flex items-center w-full gap-3 px-3 py-2.5 rounded text-sm font-medium transition-colors',
                  isUserMenuActive
                    ? 'bg-lte-primary text-white shadow-sm'
                    : 'text-lte-sidebar-text hover:bg-lte-sidebar-hover hover:text-white',
                  sidebarCollapsed ? 'lg:justify-center lg:px-2' : ''
                ]"
              >
                <img
                  :src="icons.userSquare"
                  class="size-[18px] shrink-0 brightness-0 invert"
                  :class="isUserMenuActive ? 'opacity-100' : 'opacity-80'"
                  alt=""
                >
                <span
                  v-if="!sidebarCollapsed || mobileSidebarOpen"
                  class="flex-1 text-left truncate"
                >
                  Manage User
                </span>
                <svg
                  v-if="!sidebarCollapsed || mobileSidebarOpen"
                  class="size-4 shrink-0 transition-transform duration-200"
                  :class="{ 'rotate-90': userMenuExpanded }"
                  viewBox="0 0 20 20"
                  fill="currentColor"
                >
                  <path
                    fill-rule="evenodd"
                    d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z"
                    clip-rule="evenodd"
                  />
                </svg>
              </button>

              <ul
                v-show="userMenuExpanded && (!sidebarCollapsed || mobileSidebarOpen)"
                class="mt-1 space-y-1 pl-3 bg-lte-sidebar-dark/60 rounded py-1"
              >
                <li>
                  <router-link
                    to="/users"
                    @click="closeMobileSidebar"
                    :class="[
                      'flex items-center gap-2.5 px-3 py-2 rounded text-sm transition-colors',
                      route.path === '/users' || route.path.startsWith('/users/add') || route.path.startsWith('/users/edit')
                        ? 'bg-white/95 text-lte-sidebar font-semibold'
                        : 'text-lte-sidebar-text hover:bg-lte-sidebar-hover hover:text-white'
                    ]"
                  >
                    <span class="size-2 rounded-full border border-current shrink-0"></span>
                    <span class="truncate">Users List</span>
                  </router-link>
                </li>
                <li>
                  <router-link
                    to="/users-roles"
                    @click="closeMobileSidebar"
                    :class="[
                      'flex items-center gap-2.5 px-3 py-2 rounded text-sm transition-colors',
                      route.path.startsWith('/users-roles')
                        ? 'bg-white/95 text-lte-sidebar font-semibold'
                        : 'text-lte-sidebar-text hover:bg-lte-sidebar-hover hover:text-white'
                    ]"
                  >
                    <span class="size-2 rounded-full border border-current shrink-0"></span>
                    <span class="truncate">Assign Role</span>
                  </router-link>
                </li>
              </ul>
            </li>

            <!-- Profile & Settings -->
            <li v-for="item in accountMenuItems" :key="item.to">
              <router-link
                :to="item.to"
                @click="closeMobileSidebar"
                :title="sidebarCollapsed ? item.label : ''"
                :class="[
                  'flex items-center gap-3 px-3 py-2.5 rounded text-sm font-medium transition-colors',
                  isItemActive(item)
                    ? 'bg-lte-primary text-white shadow-sm'
                    : 'text-lte-sidebar-text hover:bg-lte-sidebar-hover hover:text-white',
                  sidebarCollapsed ? 'lg:justify-center lg:px-2' : ''
                ]"
              >
                <img
                  :src="item.icon"
                  class="size-[18px] shrink-0 brightness-0 invert"
                  :class="isItemActive(item) ? 'opacity-100' : 'opacity-80'"
                  alt=""
                >
                <span v-if="!sidebarCollapsed || mobileSidebarOpen" class="truncate">
                  {{ item.label }}
                </span>
              </router-link>
            </li>
          </ul>
        </div>
      </div>

      <!-- Bottom Version Tag -->
      <div
        v-if="!sidebarCollapsed || mobileSidebarOpen"
        class="px-4 py-2.5 border-t border-lte-sidebar-border text-[11px] text-lte-sidebar-text/60 flex items-center justify-between shrink-0"
      >
        <span>CODINA Admin</span>
        <span>v1.0</span>
      </div>
    </aside>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { sidebarCollapsed, mobileSidebarOpen, closeMobileSidebar } from '@/js/layoutState'
import BrandLogo from '@/components/BrandLogo.vue'

import defaultAvatar from '@/assets/images/photos/photos-1.png'
import homeIcon from '@/assets/images/icons/home-black.svg'
import bagIcon from '@/assets/images/icons/bag-black.svg'
import noteIcon from '@/assets/images/icons/note-2-black.svg'
import buildingsIcon from '@/assets/images/icons/buildings-2-black.svg'
import shopIcon from '@/assets/images/icons/shop-black.svg'
import cardIcon from '@/assets/images/icons/card-black.svg'
import stickynoteIcon from '@/assets/images/icons/stickynote-black.svg'
import userSquareIcon from '@/assets/images/icons/user-square-black.svg'
import settingIcon from '@/assets/images/icons/setting-black.svg'
import profileIcon from '@/assets/images/icons/profile-circle-black.svg'

const props = defineProps({
  roleMode: {
    type: String,
    default: 'manager'
  }
})

const route = useRoute()
const authStore = useAuthStore()
const userMenuExpanded = ref(false)

const icons = {
  stickynote: stickynoteIcon,
  userSquare: userSquareIcon
}

const currentUser = computed(() => authStore.currentUser)

const homeRoute = computed(() =>
  props.roleMode === 'keeper' ? '/merchant-overview' : '/overview'
)

const mainMenuItems = computed(() => {
  if (props.roleMode === 'keeper') {
    return [
      { label: 'Overview', to: '/merchant-overview', icon: homeIcon, exact: true },
      { label: 'Transactions', to: '/transactions', icon: cardIcon, prefix: '/transactions' },
      { label: 'My Merchants', to: '/my-merchants', icon: shopIcon, prefix: '/my-merchants' }
    ]
  }
  return [
    { label: 'Overview', to: '/overview', icon: homeIcon, exact: true },
    { label: 'Products', to: '/products', icon: bagIcon, prefix: '/products' },
    { label: 'Categories', to: '/categories', icon: noteIcon, prefix: '/categories' },
    { label: 'Warehouses', to: '/warehouses', icon: buildingsIcon, prefix: '/warehouses' },
    { label: 'Merchants', to: '/merchants', icon: shopIcon, prefix: '/merchants' }
  ]
})

const accountMenuItems = computed(() => [
  { label: 'Profile', to: '/profile', icon: profileIcon, prefix: '/profile' },
  { label: 'Settings', to: '/settings', icon: settingIcon, prefix: '/settings' }
])

const isUserMenuActive = computed(() =>
  route.path === '/users' ||
  route.path.startsWith('/users-roles') ||
  route.path.startsWith('/users/add') ||
  route.path.startsWith('/users/edit')
)

const isItemActive = (item) => {
  if (item.exact) {
    return route.path === item.to
  }
  if (item.prefix) {
    return route.path.startsWith(item.prefix)
  }
  return route.path === item.to
}

const handleUserMenuClick = () => {
  if (sidebarCollapsed.value && !mobileSidebarOpen.value) {
    sidebarCollapsed.value = false
    userMenuExpanded.value = true
    return
  }
  userMenuExpanded.value = !userMenuExpanded.value
}

watch(
  () => route.path,
  () => {
    if (isUserMenuActive.value) {
      userMenuExpanded.value = true
    }
  }
)

onMounted(() => {
  if (isUserMenuActive.value) {
    userMenuExpanded.value = true
  }
})
</script>

<style scoped>
.lte-sidebar-scroll::-webkit-scrollbar {
  width: 5px;
}
.lte-sidebar-scroll::-webkit-scrollbar-track {
  background: transparent;
}
.lte-sidebar-scroll::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
}
.lte-sidebar-scroll::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.35);
}
</style>
