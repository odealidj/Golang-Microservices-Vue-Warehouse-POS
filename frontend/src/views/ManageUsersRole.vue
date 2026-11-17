<template>
  <Layout pageTitle="User Roles" breadcrumbText="" breadcrumbLink="">
    <section id="Products" class="lte-card-primary flex flex-col flex-1">
      <div id="Header" class="lte-card-header">
        <div>
          <h2 class="lte-card-title">
            <span>All User Roles</span>
            <span class="lte-badge bg-blue-50 text-lte-primary border border-blue-200">
              {{ totalRecords }} Total
            </span>
          </h2>
          <p class="text-xs text-lte-muted mt-0.5">View and update user role assignments here.</p>
        </div>
        <router-link to="/users-roles/assign-role" class="lte-btn-primary">
          + Assign Role
        </router-link>
      </div>

      <div id="Product-List" class="lte-card-body flex-1 flex flex-col">
        <!-- Loading State -->
        <div v-if="loading" class="flex flex-col flex-1 items-center justify-center py-12 gap-3">
          <img src="@/assets/images/icons/loading.svg" class="size-9 animate-spin" alt="loading">
          <p class="text-sm font-medium text-lte-muted">Loading user roles data...</p>
        </div>

        <!-- Error State -->
        <div v-else-if="error" class="flex flex-col flex-1 items-center justify-center py-12 gap-3 text-center">
          <p class="text-sm font-semibold text-red-600">{{ error }}</p>
          <button @click="fetchUserRoles()" class="lte-btn-primary">Try Again</button>
        </div>

        <!-- User Roles Table -->
        <div v-else-if="userRoles.length > 0" class="lte-table-wrap">
          <table class="lte-table">
            <thead>
              <tr>
                <th>User</th>
                <th>Contact</th>
                <th>Assigned Role</th>
                <th class="text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="userRole in userRoles" :key="userRole.id">
                <td>
                  <div class="flex items-center gap-3">
                    <img
                      :src="userRole.user.photo || defaultUserPhoto"
                      class="size-10 rounded-full border border-lte-border object-cover shrink-0"
                      :alt="userRole.user.name"
                      @error="setDefaultPhoto"
                    >
                    <span class="font-semibold text-lte-text whitespace-nowrap">{{ userRole.user.name }}</span>
                  </div>
                </td>
                <td class="text-lte-muted whitespace-nowrap">
                  {{ userRole.user.phone || userRole.user.email }}
                </td>
                <td>
                  <span class="lte-badge bg-blue-50 text-lte-primary border border-blue-200 capitalize">
                    {{ userRole.role.name }}
                  </span>
                </td>
                <td class="text-right">
                  <router-link :to="`/users-roles/edit-role/${userRole.id}`" class="lte-btn-warning text-xs py-1 px-2.5">
                    Edit Role
                  </router-link>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Empty State -->
        <div v-else class="flex flex-col flex-1 items-center justify-center rounded border border-dashed border-gray-300 gap-3 py-12">
          <img src="@/assets/images/icons/document-text-grey.svg" class="size-10 opacity-70" alt="empty">
          <p class="text-sm font-medium text-lte-muted">Oops, it looks like there's no data yet.</p>
        </div>
      </div>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="lte-card-footer flex items-center justify-between gap-3">
        <p class="text-xs text-lte-muted">Page {{ currentPage }} of {{ totalPages }}</p>
        <div class="flex items-center gap-1">
          <button @click="goToPage(currentPage - 1)" :disabled="!hasPrev" class="lte-btn-default text-xs py-1 px-2.5">
            Previous
          </button>
          <button
            v-for="page in totalPages"
            :key="page"
            @click="goToPage(page)"
            :class="[
              'px-2.5 py-1 rounded text-xs font-semibold transition-colors',
              currentPage === page
                ? 'bg-lte-primary text-white'
                : 'bg-white border border-lte-border text-lte-text hover:bg-gray-100'
            ]"
          >
            {{ page }}
          </button>
          <button @click="goToPage(currentPage + 1)" :disabled="!hasNext" class="lte-btn-default text-xs py-1 px-2.5">
            Next
          </button>
        </div>
      </div>
    </section>
  </Layout>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Layout from '@/components/Layout.vue'
import { getUserRoles } from '@/js/api'
import defaultUserPhoto from '@/assets/images/photos/photos-2.png'

const userRoles = ref([])
const loading = ref(false)
const error = ref(null)

const currentPage = ref(1)
const totalPages = ref(1)
const totalRecords = ref(0)
const limit = ref(10)
const hasNext = ref(false)
const hasPrev = ref(false)

const fetchUserRoles = async (page = 1) => {
  loading.value = true
  error.value = null
  try {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.value.toString()
    })
    const response = await getUserRoles(`?${params.toString()}`)
    userRoles.value = response.data.user_roles
    currentPage.value = response.data.pagination.current_page
    totalPages.value = response.data.pagination.total_pages
    totalRecords.value = response.data.pagination.total_records
    limit.value = response.data.pagination.limit
    hasNext.value = response.data.pagination.has_next
    hasPrev.value = response.data.pagination.has_prev
  } catch (err) {
    error.value = err.message || 'Error occurred while fetching user roles'
    console.error('Error fetching user roles:', err)
  } finally {
    loading.value = false
  }
}

const goToPage = (page) => {
  if (page >= 1 && page <= totalPages.value) {
    fetchUserRoles(page)
  }
}

const setDefaultPhoto = (event) => {
  event.target.src = defaultUserPhoto
}

onMounted(() => {
  fetchUserRoles()
})
</script>
