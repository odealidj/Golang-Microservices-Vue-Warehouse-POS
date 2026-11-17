<template>
  <div id="main-container" class="flex flex-1">
    <MerchantSidebar />

    <div id="Content" class="flex flex-col flex-1 p-6 pt-0 bg-lte-bg min-w-0">
      <MerchantTopBar
        title="Manage Transactions"
        breadcrumbText="Merchant Overview"
        breadcrumbLink="/merchant-overview"
      />

      <main class="flex flex-col gap-5 flex-1">
        <!-- Merchant Info Card -->
        <section id="Warehouse-Info" class="lte-card-info p-4 flex flex-wrap items-center justify-between gap-4">
          <div class="flex items-center gap-3 min-w-0">
            <img
              :src="merchantInfo.photo || defaultMerchantPhoto"
              class="size-12 rounded border border-lte-border object-contain bg-gray-50 shrink-0"
              alt="merchant"
              @error="setDefaultMerchantPhoto"
            >
            <div>
              <p class="font-semibold text-base text-lte-text">{{ merchantInfo.name }}</p>
              <p class="text-xs text-lte-muted">{{ merchantInfo.phone }}</p>
            </div>
          </div>

          <div class="text-left sm:text-right">
            <p class="text-xs text-lte-muted">Keeper Name</p>
            <p class="text-sm font-semibold text-lte-text">{{ merchantInfo.keeper_name }} (You)</p>
          </div>
        </section>

        <!-- Transactions Card -->
        <section id="Products" class="lte-card-primary flex flex-col flex-1">
          <div id="Header" class="lte-card-header">
            <div>
              <h2 class="lte-card-title">
                <span>All Transactions</span>
                <span class="lte-badge bg-blue-50 text-lte-primary border border-blue-200">
                  {{ pagination.total_records }} Total
                </span>
              </h2>
              <p class="text-xs text-lte-muted mt-0.5">View and update your transactions list here.</p>
            </div>
            <div class="flex items-center gap-2">
              <button @click="fetchTransactions" class="lte-btn-default">Refresh</button>
              <router-link to="/transactions/add-1" class="lte-btn-primary">+ Add New</router-link>
            </div>
          </div>

          <div id="Product-List" class="lte-card-body flex-1 flex flex-col gap-4">
            <div v-if="transactions.length === 0" class="flex flex-col flex-1 items-center justify-center rounded border border-dashed border-gray-300 gap-3 py-12">
              <img src="@/assets/images/icons/document-text-grey.svg" class="size-10 opacity-70" alt="empty">
              <p class="text-sm font-medium text-lte-muted">Oops, it looks like there's no data yet.</p>
            </div>

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
                  <span
                    v-if="transaction.payment_status"
                    class="lte-badge capitalize"
                    :class="statusBadgeClass(transaction.payment_status)"
                  >
                    {{ transaction.payment_status }}
                  </span>
                  <span class="text-sm font-bold text-lte-primary">
                    Grandtotal: Rp {{ formatNumber(transaction.grand_total) }}
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
                              :src="product.product_photo || galleryDefaultIcon"
                              class="size-9 rounded border border-lte-border object-contain bg-gray-50 shrink-0"
                              alt="product"
                              @error="setDefaultProductPhoto"
                            >
                            <span class="font-medium text-lte-text">{{ product.product_name }}</span>
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
        <div class="p-5 space-y-4">
          <div class="flex items-start justify-between gap-4">
            <div class="space-y-1">
              <span class="lte-badge bg-gray-100 text-lte-muted">
                {{ selectedProduct.category?.name || 'Uncategorized' }}
              </span>
              <p class="font-bold text-base text-lte-text">{{ selectedProduct.product_name }}</p>
              <p class="font-semibold text-sm text-lte-primary">Rp {{ formatNumber(selectedProduct.price) }}</p>
            </div>
            <img
              :src="selectedProduct.product_photo || galleryDefaultIcon"
              class="size-20 rounded border border-lte-border object-contain bg-gray-50 shrink-0"
              alt="product"
              @error="setDefaultProductPhoto"
            >
          </div>
          <div class="border-t border-lte-border pt-3">
            <p class="text-xs font-semibold uppercase text-lte-muted mb-1">Product About</p>
            <p class="text-sm text-lte-text leading-relaxed">{{ selectedProduct.product_about || '-' }}</p>
          </div>
          <div class="border-t border-lte-border pt-3 space-y-1.5 text-sm">
            <div class="flex items-center justify-between">
              <span class="text-lte-muted">Quantity</span>
              <span class="font-semibold text-lte-text">{{ selectedProduct.quantity }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-lte-muted">Sub Total</span>
              <span class="font-bold text-lte-primary">Rp {{ formatNumber(selectedProduct.sub_total) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { getTransactions } from '@/js/api'
import { formatNumber, galleryDefaultIcon } from '@/js/formatters'
import defaultMerchantPhoto from '@/assets/images/thumbnails/merchant-1.png'
import MerchantSidebar from '@/components/MerchantSidebar.vue'
import MerchantTopBar from '@/components/MerchantTopBar.vue'

const authStore = useAuthStore()

const transactions = ref([])
const pagination = ref({
  current_page: 1,
  total_pages: 1,
  total_records: 0,
  limit: 10,
  has_next: false,
  has_prev: false
})

const expandedSections = ref([])
const showModal = ref(false)
const selectedProduct = ref({})

const merchantData = computed(() => authStore.getMerchantData())

const merchantInfo = computed(() => {
  if (merchantData.value && merchantData.value.length > 0) {
    return merchantData.value[0]
  }
  return {
    name: 'Mazdo Merchant',
    phone: '0812923920329',
    keeper_name: 'Raratih Aqilla',
    photo: defaultMerchantPhoto
  }
})

const statusBadgeClass = (status) => {
  switch ((status || '').toLowerCase()) {
    case 'completed':
    case 'settlement':
    case 'paid':
      return 'bg-green-100 text-green-800'
    case 'pending':
      return 'bg-yellow-100 text-yellow-800'
    case 'failed':
    case 'cancelled':
      return 'bg-red-100 text-red-800'
    default:
      return 'bg-gray-100 text-lte-muted'
  }
}

const fetchTransactions = async () => {
  try {
    const merchantId = merchantInfo.value?.id
    const query = merchantId ? `?merchant_id=${merchantId}` : ''
    const response = await getTransactions(query)

    if (response && response.data) {
      transactions.value = response.data.transactions || []
      pagination.value = response.data.pagination || {
        current_page: 1,
        total_pages: 1,
        total_records: transactions.value.length,
        limit: 10,
        has_next: false,
        has_prev: false
      }
    } else if (response && response.transactions) {
      transactions.value = response.transactions || []
      pagination.value = response.pagination || {
        current_page: 1,
        total_pages: 1,
        total_records: transactions.value.length,
        limit: 10,
        has_next: false,
        has_prev: false
      }
    } else {
      transactions.value = []
    }
  } catch (err) {
    console.error('Error fetching transactions:', err)
    transactions.value = []
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
}

const setDefaultProductPhoto = (event) => {
  event.target.src = galleryDefaultIcon
}

const setDefaultMerchantPhoto = (event) => {
  event.target.src = defaultMerchantPhoto
}

onMounted(async () => {
  if (!merchantData.value && authStore.userRole?.toLowerCase() === 'keeper') {
    await authStore.refreshMerchantData()
  }
  await new Promise((resolve) => setTimeout(resolve, 100))
  await fetchTransactions()
})
</script>
