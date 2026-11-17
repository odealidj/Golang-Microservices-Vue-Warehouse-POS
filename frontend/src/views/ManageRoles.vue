<template>
  <Layout pageTitle="Manage Roles" breadcrumbText="" breadcrumbLink="">
    <main class="flex flex-col flex-1">
      <section id="Products" class="lte-card-primary flex flex-col flex-1">
        <div id="Header" class="lte-card-header">
          <div>
            <h2 class="lte-card-title">
              <span>All Roles</span>
              <span class="lte-badge bg-blue-50 text-lte-primary border border-blue-200">
                {{ roles.length }} Total
              </span>
            </h2>
            <p class="text-xs text-lte-muted mt-0.5">View and update your Roles here.</p>
          </div>
          <router-link to="/roles/add" class="lte-btn-primary">
            + Add New
          </router-link>
        </div>

        <div id="Product-List" class="lte-card-body flex-1 flex flex-col">
          <!-- Loading State -->
          <div v-if="loading" class="flex flex-col flex-1 items-center justify-center py-12 gap-3">
            <img src="@/assets/images/icons/loading.svg" class="size-9 animate-spin" alt="loading">
            <p class="text-sm font-medium text-lte-muted">Loading roles...</p>
          </div>

          <!-- Error State -->
          <div v-else-if="error" class="flex flex-col flex-1 items-center justify-center py-12 gap-3 text-center">
            <p class="text-sm font-semibold text-red-600">{{ error }}</p>
            <button @click="rolesStore.fetchRoles()" class="lte-btn-primary">Try Again</button>
          </div>

          <!-- Empty State -->
          <div v-else-if="roles.length === 0" id="Empty-State" class="flex flex-col flex-1 items-center justify-center rounded border border-dashed border-gray-300 gap-3 py-12">
            <img src="@/assets/images/icons/document-text-grey.svg" class="size-10 opacity-70" alt="empty">
            <p class="text-sm font-medium text-lte-muted">Oops, it looks like there's no data yet.</p>
          </div>

          <!-- Roles Data Table -->
          <div v-else class="lte-table-wrap">
            <table class="lte-table">
              <thead>
                <tr>
                  <th>Role Name</th>
                  <th>Assigned Users</th>
                  <th class="text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="role in roles" :key="role.id">
                  <td>
                    <div class="flex items-center gap-3">
                      <div class="size-9 rounded bg-gray-100 border border-lte-border flex items-center justify-center shrink-0">
                        <img src="@/assets/images/icons/user-octagon-grey.svg" class="size-5" alt="role">
                      </div>
                      <span class="font-semibold text-lte-text capitalize">{{ role.name }}</span>
                    </div>
                  </td>
                  <td>
                    <span class="lte-badge bg-gray-100 text-lte-text">
                      {{ role.count_users }} Total User
                    </span>
                  </td>
                  <td class="text-right">
                    <router-link :to="`/roles/edit/${role.id}`" class="lte-btn-warning text-xs py-1 px-2.5">
                      Edit
                    </router-link>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>
    </main>
  </Layout>
</template>

<script setup>
import { onMounted, computed } from 'vue'
import Layout from '@/components/Layout.vue'
import { useRolesStore } from '@/stores/roles'

const rolesStore = useRolesStore()

const roles = computed(() => rolesStore.roles)
const loading = computed(() => rolesStore.loading)
const error = computed(() => rolesStore.error)

onMounted(() => {
  rolesStore.fetchRoles()
})
</script>
