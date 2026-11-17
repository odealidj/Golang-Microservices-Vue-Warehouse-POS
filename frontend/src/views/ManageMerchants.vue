<template>
  <Layout page-title="Manage Merchants">
    <main class="flex flex-col flex-1">
      <section id="Merchants" class="lte-card-primary flex flex-col flex-1">
        <div id="Header" class="lte-card-header">
          <div>
            <h2 class="lte-card-title">
              <span>All Merchants</span>
              <span class="lte-badge bg-blue-50 text-lte-primary border border-blue-200">
                {{ merchants.length }} Total
              </span>
            </h2>
            <p class="text-xs text-lte-muted mt-0.5">View and update your Merchants list here.</p>
          </div>
          <div class="flex items-center gap-2">
            <button @click="loadMerchants" :disabled="isLoading" class="lte-btn-default">
              Refresh
            </button>
            <router-link to="/merchants/add" class="lte-btn-primary">
              + Add New
            </router-link>
          </div>
        </div>

        <div id="Merchant-List" class="lte-card-body flex-1 flex flex-col">
          <!-- Loading State -->
          <div v-if="isLoading" class="flex flex-col flex-1 items-center justify-center py-12 gap-3">
            <img src="@/assets/images/icons/loading.svg" class="size-9 animate-spin" alt="loading">
            <p class="text-sm font-medium text-lte-muted">Loading merchants...</p>
          </div>

          <!-- Empty State -->
          <div v-else-if="merchants.length === 0" class="flex flex-col flex-1 items-center justify-center rounded border border-dashed border-gray-300 gap-3 py-12">
            <img src="@/assets/images/icons/document-text-grey.svg" class="size-10 opacity-70" alt="empty">
            <p class="text-sm font-medium text-lte-muted">Oops, it looks like there's no data yet.</p>
          </div>

          <!-- Merchants Data Table -->
          <div v-else class="lte-table-wrap">
            <table class="lte-table">
              <thead>
                <tr>
                  <th>Merchant</th>
                  <th>Address</th>
                  <th>Assigned Products</th>
                  <th class="text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="merchant in paginatedMerchants" :key="merchant.id">
                  <td>
                    <div class="flex items-center gap-3">
                      <img
                        :src="getMerchantImage(merchant.photo)"
                        class="size-11 rounded border border-lte-border object-contain bg-gray-50 shrink-0"
                        :alt="merchant.name"
                      >
                      <span class="font-semibold text-lte-text whitespace-nowrap">{{ merchant.name }}</span>
                    </div>
                  </td>
                  <td class="text-lte-muted max-w-[260px] truncate">{{ merchant.address }}</td>
                  <td class="whitespace-nowrap">
                    <span class="lte-badge bg-gray-100 text-lte-text">
                      {{ formatNumber(merchant.product_count) }} Products
                    </span>
                  </td>
                  <td class="text-right whitespace-nowrap">
                    <div class="inline-flex items-center gap-1.5">
                      <router-link :to="`/merchants/${merchant.id}`" class="lte-btn-info text-xs py-1 px-2.5">
                        Details
                      </router-link>
                      <router-link :to="`/merchants/edit/${merchant.id}`" class="lte-btn-warning text-xs py-1 px-2.5">
                        Edit
                      </router-link>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Pagination -->
        <div v-if="merchants.length > 0" class="lte-card-footer flex flex-wrap items-center justify-between gap-3">
          <p class="text-xs text-lte-muted">
            Showing {{ startIndex + 1 }}-{{ endIndex }} of {{ merchants.length }} merchants
          </p>
          <div class="flex items-center gap-1">
            <button @click="previousPage" :disabled="currentPage === 1" class="lte-btn-default text-xs py-1 px-2.5">
              Previous
            </button>
            <button
              v-for="page in visiblePages"
              :key="page"
              @click="goToPage(page)"
              :class="[
                'px-2.5 py-1 rounded text-xs font-semibold transition-colors',
                page === currentPage
                  ? 'bg-lte-primary text-white'
                  : 'bg-white border border-lte-border text-lte-text hover:bg-gray-100'
              ]"
            >
              {{ page }}
            </button>
            <button @click="nextPage" :disabled="currentPage === totalPages" class="lte-btn-default text-xs py-1 px-2.5">
              Next
            </button>
          </div>
        </div>
      </section>
    </main>
  </Layout>
</template>

<script>
import Layout from '@/components/Layout.vue'
import { getMerchants } from '@/js/api'
import shopIcon from '@/assets/images/icons/shop-black.svg'

export default {
  name: 'ManageMerchants',
  components: {
    Layout
  },
  data() {
    return {
      merchants: [],
      currentPage: 1,
      itemsPerPage: 8,
      pagination: {
        current_page: 1,
        total_pages: 1,
        total_records: 0,
        limit: 10,
        has_next: false,
        has_prev: false
      },
      isLoading: false
    }
  },
  computed: {
    totalPages() {
      return Math.max(1, Math.ceil(this.merchants.length / this.itemsPerPage))
    },
    startIndex() {
      return (this.currentPage - 1) * this.itemsPerPage
    },
    endIndex() {
      return Math.min(this.startIndex + this.itemsPerPage, this.merchants.length)
    },
    paginatedMerchants() {
      return this.merchants.slice(this.startIndex, this.endIndex)
    },
    visiblePages() {
      const pages = []
      const maxVisible = 5
      let start = Math.max(1, this.currentPage - Math.floor(maxVisible / 2))
      const end = Math.min(this.totalPages, start + maxVisible - 1)
      if (end - start + 1 < maxVisible) {
        start = Math.max(1, end - maxVisible + 1)
      }
      for (let i = start; i <= end; i++) {
        pages.push(i)
      }
      return pages
    }
  },
  async created() {
    await this.loadMerchants()
  },
  methods: {
    async loadMerchants() {
      this.isLoading = true
      try {
        const response = await getMerchants()
        const payload = response.data
        this.merchants = Array.isArray(payload) ? payload : (payload?.data || [])
        this.pagination = response.pagination || payload?.pagination || {
          current_page: 1,
          total_pages: 1,
          total_records: this.merchants.length,
          limit: 10,
          has_next: false,
          has_prev: false
        }
      } catch (error) {
        console.error('Error loading merchants:', error)
        this.merchants = []
      } finally {
        this.isLoading = false
      }
    },
    goToPage(page) {
      this.currentPage = page
    },
    nextPage() {
      if (this.currentPage < this.totalPages) this.currentPage++
    },
    previousPage() {
      if (this.currentPage > 1) this.currentPage--
    },
    getMerchantImage(photo) {
      return photo || shopIcon
    },
    formatNumber(num) {
      return new Intl.NumberFormat('en-US').format(num || 0)
    }
  }
}
</script>
