<template>
  <div id="main-container" class="flex flex-1">
    <!-- Sidebar -->
    <MerchantSidebar />

    <!-- Main Content -->
    <div id="Content" class="flex flex-col flex-1 p-6 pt-0 bg-lte-bg min-w-0">
      <!-- Top Bar -->
      <MerchantTopBar
        title="Merchant Overview"
        breadcrumbText="Back to Transactions"
        breadcrumbLink="/transactions"
      />

      <!-- Main Content -->
      <main class="flex flex-col gap-5 flex-1">
        <!-- Quick Access Toolbar -->
        <div class="flex flex-wrap items-center justify-between gap-3">
          <router-link to="/transactions/add-1" class="lte-btn-primary">
            <span>+ Add New Transaction</span>
          </router-link>

          <button
            @click="refreshDashboard"
            class="lte-btn-default"
            :disabled="dashboardLoading"
          >
            <img src="@/assets/images/icons/loading.svg" class="size-4 shrink-0" :class="{ 'animate-spin': dashboardLoading }" alt="refresh">
            <span>Refresh Dashboard</span>
          </button>
        </div>

        <!-- AdminLTE Small Boxes (KPI) -->
        <section class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div v-if="dashboardLoading" class="md:col-span-3 lte-card p-10 flex flex-col items-center justify-center">
            <img src="@/assets/images/icons/loading.svg" class="size-8 animate-spin mb-2" alt="loading">
            <p class="text-sm text-lte-muted">Loading dashboard data...</p>
          </div>

          <div v-else-if="!currentMerchant" class="md:col-span-3 lte-card p-10 flex flex-col items-center justify-center text-center">
            <img src="@/assets/images/icons/shop-grey.svg" class="size-10 mb-2 opacity-70" alt="no merchant">
            <p class="text-sm text-lte-muted">No merchant data found. Please contact administrator.</p>
          </div>

          <template v-else>
            <div id="Total-Revenue" class="rounded bg-lte-info text-white shadow-sm overflow-hidden flex flex-col justify-between">
              <div class="p-4 flex items-center justify-between gap-3">
                <div class="min-w-0">
                  <p class="text-2xl lg:text-[28px] font-bold truncate">Rp {{ formatCurrency(dashboardData.total_revenue) }}</p>
                  <p class="text-sm text-white/90 mt-1">Total Revenue</p>
                </div>
                <img src="@/assets/images/icons/wallet-blue-fill.svg" class="size-12 opacity-30 brightness-0 invert shrink-0" alt="revenue">
              </div>
              <router-link to="/transactions" class="px-4 py-1.5 bg-black/15 hover:bg-black/25 text-xs font-medium text-white/95 flex items-center justify-center gap-1 transition-colors">
                <span>View Transactions</span>
                <span>&rarr;</span>
              </router-link>
            </div>

            <div id="Total-Transactions" class="rounded bg-lte-success text-white shadow-sm overflow-hidden flex flex-col justify-between">
              <div class="p-4 flex items-center justify-between gap-3">
                <div class="min-w-0">
                  <p class="text-2xl lg:text-[28px] font-bold">{{ formatNumber(dashboardData.total_transactions) }}</p>
                  <p class="text-sm text-white/90 mt-1">Total Transactions</p>
                </div>
                <img src="@/assets/images/icons/document-text-blue-fill.svg" class="size-12 opacity-30 brightness-0 invert shrink-0" alt="transactions">
              </div>
              <router-link to="/transactions" class="px-4 py-1.5 bg-black/15 hover:bg-black/25 text-xs font-medium text-white/95 flex items-center justify-center gap-1 transition-colors">
                <span>View Transactions</span>
                <span>&rarr;</span>
              </router-link>
            </div>

            <div id="Products-Sold" class="rounded bg-lte-warning text-gray-900 shadow-sm overflow-hidden flex flex-col justify-between">
              <div class="p-4 flex items-center justify-between gap-3">
                <div class="min-w-0">
                  <p class="text-2xl lg:text-[28px] font-bold">{{ formatNumber(dashboardData.products_sold) }}</p>
                  <p class="text-sm text-gray-900/80 mt-1">Products Sold</p>
                </div>
                <img src="@/assets/images/icons/bag-blue-fill.svg" class="size-12 opacity-25 brightness-0 shrink-0" alt="products">
              </div>
              <router-link to="/my-merchants" class="px-4 py-1.5 bg-black/10 hover:bg-black/15 text-xs font-medium text-gray-900 flex items-center justify-center gap-1 transition-colors">
                <span>My Merchants</span>
                <span>&rarr;</span>
              </router-link>
            </div>
          </template>
        </section>

        <!-- Latest Transactions Card -->
        <section id="Latest-Transaction" class="lte-card-primary flex-1">
          <div class="lte-card-header">
            <h2 class="lte-card-title">Latest Transaction</h2>
            <div v-if="loading" class="flex items-center gap-2 text-xs text-lte-muted">
              <img src="@/assets/images/icons/loading.svg" class="size-4 animate-spin" alt="loading">
              <span>Loading...</span>
            </div>
          </div>

          <div class="lte-card-body">
            <div v-if="loading && transactions.length === 0" class="flex flex-col items-center justify-center py-10">
              <img src="@/assets/images/icons/loading.svg" class="size-8 animate-spin mb-2" alt="loading">
              <p class="text-sm text-lte-muted">Loading transactions...</p>
            </div>

            <div v-else-if="transactions.length > 0" class="space-y-4">
              <div
                v-for="(transaction, index) in transactions"
                :key="transaction.id"
                class="rounded border border-lte-border overflow-hidden"
              >
                <div class="flex flex-wrap items-center justify-between gap-3 px-4 py-3 bg-gray-50 border-b border-lte-border">
                  <div class="flex items-center gap-3 min-w-0">
                    <div class="size-10 rounded bg-white border border-lte-border flex items-center justify-center shrink-0">
                      <img src="@/assets/images/icons/user-thin-grey.svg" class="size-5" alt="customer">
                    </div>
                    <div class="min-w-0">
                      <p class="font-semibold text-sm text-lte-text truncate">{{ transaction.name }}</p>
                      <p class="text-xs text-lte-muted">{{ transaction.phone }}</p>
                    </div>
                  </div>
                  <div class="flex items-center gap-3">
                    <span class="lte-badge bg-blue-50 text-lte-primary border border-blue-200">
                      {{ transaction.merchant_name }}
                    </span>
                    <span class="text-sm font-bold text-lte-primary">
                      Rp {{ formatNumber(transaction.grand_total) }}
                    </span>
                  </div>
                </div>

                <div class="px-4 py-2.5 bg-white">
                  <button
                    @click="toggleProductAssigned(index + 1)"
                    type="button"
                    class="flex items-center justify-between w-full text-xs font-semibold text-lte-muted hover:text-lte-text py-1"
                  >
                    <span>Product Assigned ({{ transaction.transaction_products?.length || 0 }})</span>
                    <svg
                      class="size-4 transition-transform duration-200"
                      :class="{ 'rotate-180': expandedSections.includes(index + 1) }"
                      viewBox="0 0 20 20"
                      fill="currentColor"
                    >
                      <path
                        fill-rule="evenodd"
                        d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z"
                        clip-rule="evenodd"
                      />
                    </svg>
                  </button>

                  <div v-show="expandedSections.includes(index + 1)" class="mt-2 lte-table-wrap border-t border-lte-border pt-2">
                    <table class="lte-table">
                      <thead>
                        <tr>
                          <th>Product</th>
                          <th>Category</th>
                          <th>Price &times; Qty</th>
                          <th class="text-right">Action</th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr v-for="product in transaction.transaction_products" :key="product.id">
                          <td>
                            <div class="flex items-center gap-2.5">
                              <img
                                :src="product.product_photo"
                                class="size-9 rounded border border-lte-border object-contain bg-gray-50 shrink-0"
                                alt="product"
                                @error="setDefaultProductPhoto"
                              >
                              <span class="font-medium text-lte-text line-clamp-1">{{ product.product_name }}</span>
                            </div>
                          </td>
                          <td class="text-lte-muted whitespace-nowrap">
                            {{ product.category?.name || 'Uncategorized' }}
                          </td>
                          <td class="whitespace-nowrap font-medium text-lte-primary">
                            Rp {{ formatNumber(product.price) }}
                            <span class="text-lte-muted">({{ product.quantity }}x)</span>
                          </td>
                          <td class="text-right">
                            <button @click="showProductDetails(product)" class="lte-btn-default text-xs py-1 px-2.5">
                              Details
                            </button>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                </div>
              </div>
            </div>

            <div v-else-if="!loading" class="flex flex-col items-center justify-center rounded border border-dashed border-gray-300 gap-3 py-12 text-center">
              <img src="@/assets/images/icons/document-text-grey.svg" class="size-10 opacity-70" alt="empty">
              <p class="text-sm font-medium text-lte-muted">Oops, it looks like there's no transaction data yet.</p>
            </div>
          </div>
        </section>
      </main>
    </div>

    <!-- Product Details Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div @click="closeModal" class="absolute inset-0 bg-black/50"></div>
      <div class="relative bg-white rounded border border-lte-border shadow-xl w-full max-w-md overflow-hidden">
        <div class="lte-card-header bg-gray-50">
          <h3 class="lte-card-title">Product Details</h3>
          <button @click="closeModal" class="text-gray-400 hover:text-gray-600 text-xl leading-none">&times;</button>
        </div>
        <div class="p-5 space-y-4" v-if="selectedProduct">
          <div class="flex items-start justify-between gap-4">
            <div class="space-y-1">
              <span class="lte-badge bg-gray-100 text-lte-muted">
                {{ selectedProduct.category?.name || 'Uncategorized' }}
              </span>
              <p class="font-bold text-base text-lte-text">{{ selectedProduct.product_name }}</p>
              <p class="font-semibold text-sm text-lte-primary">
                Rp {{ formatNumber(selectedProduct.price) }}
                <span class="text-lte-muted font-normal">(Qty: {{ selectedProduct.quantity }})</span>
              </p>
            </div>
            <img
              :src="selectedProduct.product_photo"
              class="size-20 rounded border border-lte-border object-contain bg-gray-50 shrink-0"
              alt="product"
              @error="setDefaultProductPhoto"
            >
          </div>
          <div class="border-t border-lte-border pt-3">
            <p class="text-xs font-semibold uppercase text-lte-muted mb-1">Product About</p>
            <p class="text-sm text-lte-text leading-relaxed">{{ selectedProduct.product_about || 'No description available for this product.' }}</p>
          </div>
          <div class="border-t border-lte-border pt-3 flex items-center justify-between">
            <span class="text-sm text-lte-muted">Subtotal</span>
            <span class="text-base font-bold text-lte-primary">Rp {{ formatNumber(selectedProduct.sub_total) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import MerchantSidebar from '@/components/MerchantSidebar.vue'
import MerchantTopBar from '@/components/MerchantTopBar.vue'
import { getKeeperDashboardData, getTransactions, getFirstMerchantFromStorage } from '@/js/api'
import { formatCurrency, formatNumber, galleryDefaultIcon } from '@/js/formatters'

const transactions = ref([])
const dashboardData = ref({
  total_revenue: 0,
  total_transactions: 0,
  products_sold: 0,
  merchant: null
})
const expandedSections = ref([1, 2])
const showModal = ref(false)
const loading = ref(false)
const dashboardLoading = ref(false)
const selectedProduct = ref(null)
const currentMerchant = ref(null)
const retryCount = ref(0)
const maxRetries = 3
const refreshInterval = ref(null)

const loadDashboardData = async () => {
  try {
    dashboardLoading.value = true
    const merchant = getFirstMerchantFromStorage()
    if (!merchant) {
      window.location.href = '/'
      return
    }

    currentMerchant.value = merchant
    const response = await getKeeperDashboardData(merchant.id)
    if (response.data) {
      dashboardData.value = {
        total_revenue: response.data.total_revenue || 0,
        total_transactions: response.data.total_transactions || 0,
        products_sold: response.data.products_sold || 0,
        merchant: response.data.merchant || null
      }
    }
    retryCount.value = 0
  } catch (error) {
    console.error('Error loading dashboard data:', error)
    if (retryCount.value < maxRetries) {
      retryCount.value++
      setTimeout(() => {
        loadDashboardData()
      }, 1000 * retryCount.value)
      return
    }
    const merchant = getFirstMerchantFromStorage()
    if (merchant) {
      dashboardData.value = {
        total_revenue: 0,
        total_transactions: 0,
        products_sold: 0,
        merchant: {
          id: merchant.id,
          name: merchant.name,
          address: merchant.address,
          photo: merchant.photo
        }
      }
    } else {
      dashboardData.value = {
        total_revenue: 0,
        total_transactions: 0,
        products_sold: 0,
        merchant: null
      }
    }
  } finally {
    dashboardLoading.value = false
  }
}

const loadLatestTransactions = async () => {
  try {
    loading.value = true
    let query = '?limit=2'
    if (currentMerchant.value) {
      query += `&merchant_id=${currentMerchant.value.id}`
    }
    const response = await getTransactions(query)
    transactions.value = response.data?.transactions || response.transactions || []
  } catch (error) {
    console.error('Error loading latest transactions:', error)
    transactions.value = []
  } finally {
    loading.value = false
  }
}

const toggleProductAssigned = (sectionId) => {
  const index = expandedSections.value.indexOf(sectionId)
  if (index > -1) {
    expandedSections.value.splice(index, 1)
  } else {
    expandedSections.value.push(sectionId)
  }
}

const showProductDetails = (product) => {
  selectedProduct.value = product
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
  selectedProduct.value = null
}

const setDefaultProductPhoto = (event) => {
  event.target.src = galleryDefaultIcon
}

const refreshDashboard = async () => {
  try {
    await loadDashboardData()
    await loadLatestTransactions()
  } catch (error) {
    console.error('Error refreshing dashboard:', error)
  }
}

onMounted(async () => {
  await loadDashboardData()
  await loadLatestTransactions()

  refreshInterval.value = setInterval(async () => {
    if (!dashboardLoading.value) {
      await loadDashboardData()
    }
  }, 5 * 60 * 1000)
})

onUnmounted(() => {
  if (refreshInterval.value) {
    clearInterval(refreshInterval.value)
  }
})
</script>
