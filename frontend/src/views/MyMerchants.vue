<template>
  <div id="main-container" class="flex flex-1">
    <MerchantSidebar />

    <div id="Content" class="flex flex-col flex-1 p-6 pt-0 bg-lte-bg min-w-0">
      <MerchantTopBar
        title="My Merchants"
        breadcrumbText="Merchant Overview"
        breadcrumbLink="/merchant-overview"
      />

      <main class="flex flex-col gap-5 flex-1">
        <!-- Loading State -->
        <div v-if="loading" class="lte-card p-12 flex items-center justify-center gap-3">
          <img src="@/assets/images/icons/loading.svg" class="size-6 animate-spin" alt="loading">
          <span class="text-sm font-medium text-lte-muted">Loading merchants...</span>
        </div>

        <!-- Merchant Info Card -->
        <section v-else-if="selectedMerchant" id="Merchant-Info" class="lte-card-info p-4 flex flex-wrap items-center justify-between gap-4">
          <div class="flex items-center gap-3 min-w-0">
            <img
              :src="selectedMerchant.photo || defaultMerchantPhoto"
              class="size-12 rounded border border-lte-border object-contain bg-gray-50 shrink-0"
              alt="merchant"
              @error="setDefaultMerchantPhoto"
            >
            <div>
              <p class="font-semibold text-base text-lte-text">{{ selectedMerchant.name }}</p>
              <p class="text-xs text-lte-muted">{{ selectedMerchant.phone }}</p>
            </div>
          </div>

          <div class="flex items-center gap-4">
            <div v-if="merchants.length > 1" class="flex items-center gap-2">
              <label class="text-xs text-lte-muted font-medium">Switch Outlet:</label>
              <select
                :value="selectedMerchant.id"
                @change="selectMerchant(merchants.find(m => m.id === Number($event.target.value)))"
                class="lte-input w-48"
              >
                <option v-for="m in merchants" :key="m.id" :value="m.id">{{ m.name }}</option>
              </select>
            </div>
            <div class="text-left sm:text-right">
              <p class="text-xs text-lte-muted">Keeper Name</p>
              <p class="text-sm font-semibold text-lte-text">{{ selectedMerchant.keeper_name }} (You)</p>
            </div>
          </div>
        </section>

        <!-- Merchant Selection -->
        <section v-else-if="merchants.length > 0" id="Merchant-Selection" class="lte-card-info">
          <div class="lte-card-header">
            <h2 class="lte-card-title">Select Merchant</h2>
          </div>
          <div class="lte-card-body flex flex-wrap gap-3">
            <button
              v-for="merchant in merchants"
              :key="merchant.id"
              @click="selectMerchant(merchant)"
              class="flex items-center gap-3 px-4 py-3 rounded border border-lte-border hover:border-lte-primary hover:bg-blue-50/40 transition-colors"
            >
              <img :src="merchant.photo || defaultMerchantPhoto" class="size-10 rounded object-contain bg-gray-50" alt="merchant">
              <div class="text-left">
                <p class="font-semibold text-sm text-lte-text">{{ merchant.name }}</p>
                <p class="text-xs text-lte-muted">{{ merchant.merchant_products?.length || 0 }} Products</p>
              </div>
            </button>
          </div>
        </section>

        <!-- Products Data Table -->
        <section v-if="selectedMerchant" id="Products" class="lte-card-primary flex flex-col flex-1">
          <div id="Header" class="lte-card-header">
            <div>
              <h2 class="lte-card-title">
                <span>Outlet Products</span>
                <span class="lte-badge bg-blue-50 text-lte-primary border border-blue-200">
                  {{ merchantProducts?.length || 0 }} Total
                </span>
              </h2>
              <p class="text-xs text-lte-muted mt-0.5">View your assigned merchant products and stock levels.</p>
            </div>
            <button @click="refreshProducts" :disabled="loadingProducts" class="lte-btn-default">
              {{ loadingProducts ? 'Refreshing...' : 'Refresh' }}
            </button>
          </div>

          <div id="Product-List" class="lte-card-body flex-1 flex flex-col">
            <div v-if="loadingProducts" class="flex items-center justify-center py-12 gap-3">
              <img src="@/assets/images/icons/loading.svg" class="size-6 animate-spin" alt="loading">
              <span class="text-sm font-medium text-lte-muted">Loading products...</span>
            </div>

            <div v-else-if="merchantProducts && merchantProducts.length > 0" class="lte-table-wrap">
              <table class="lte-table">
                <thead>
                  <tr>
                    <th>Product</th>
                    <th>Category</th>
                    <th>Price</th>
                    <th>Stock</th>
                    <th class="text-right">Action</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="product in merchantProducts" :key="product.id">
                    <td>
                      <div class="flex items-center gap-3">
                        <img
                          :src="product.product_photo || galleryDefaultIcon"
                          class="size-10 rounded border border-lte-border object-contain bg-gray-50 shrink-0"
                          alt="product"
                          @error="setDefaultProductPhoto"
                        >
                        <span class="font-semibold text-lte-text max-w-[240px] truncate">{{ product.product_name }}</span>
                      </div>
                    </td>
                    <td class="text-lte-muted whitespace-nowrap">{{ product.product_category || 'Uncategorized' }}</td>
                    <td class="font-semibold text-lte-primary whitespace-nowrap">Rp {{ formatPrice(product.product_price) }}</td>
                    <td>
                      <span
                        class="lte-badge"
                        :class="product.stock < 10 ? 'bg-red-100 text-red-800' : 'bg-green-100 text-green-800'"
                      >
                        {{ product.stock }} Stock
                      </span>
                    </td>
                    <td class="text-right">
                      <button @click="showProductDetails(product)" class="lte-btn-info text-xs py-1 px-2.5">
                        Details
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div v-else class="flex flex-col flex-1 items-center justify-center rounded border border-dashed border-gray-300 gap-3 py-12">
              <img src="@/assets/images/icons/document-text-grey.svg" class="size-10 opacity-70" alt="empty">
              <p class="text-sm font-medium text-lte-muted">Oops, it looks like there's no data yet.</p>
            </div>
          </div>
        </section>

        <!-- Empty State for No Merchants -->
        <div v-else-if="showEmptyState" class="lte-card p-12 flex flex-col items-center justify-center gap-3 text-center">
          <img src="@/assets/images/icons/document-text-grey.svg" class="size-10 opacity-70" alt="empty">
          <p class="text-sm font-medium text-lte-muted">Oops, it looks like there's no merchant assigned to you yet.</p>
        </div>
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
              <span class="lte-badge bg-gray-100 text-lte-muted">{{ selectedProduct.product_category }}</span>
              <p class="font-bold text-base text-lte-text">{{ selectedProduct.product_name }}</p>
              <p class="font-semibold text-sm text-lte-primary">Rp {{ formatPrice(selectedProduct.product_price) }}</p>
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
              <span class="text-lte-muted">Stock</span>
              <span class="font-semibold text-lte-text">{{ selectedProduct.stock }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-lte-muted">Warehouse</span>
              <span class="font-semibold text-lte-text">{{ selectedProduct.warehouse_name }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-lte-muted">Warehouse Phone</span>
              <span class="font-semibold text-lte-text">{{ selectedProduct.warehouse_phone }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { getMerchants, getMerchantProducts } from '@/js/api'
import { galleryDefaultIcon } from '@/js/formatters'
import defaultMerchantPhoto from '@/assets/images/thumbnails/merchant-1.png'
import MerchantSidebar from '@/components/MerchantSidebar.vue'
import MerchantTopBar from '@/components/MerchantTopBar.vue'

const authStore = useAuthStore()

const merchants = ref([])
const selectedMerchant = ref(null)
const merchantProducts = ref([])
const loading = ref(true)
const loadingProducts = ref(false)
const showModal = ref(false)
const showEmptyState = ref(false)
const selectedProduct = ref({})

const fetchMerchants = async () => {
  try {
    loading.value = true
    const storedMerchantData = authStore.getMerchantData()

    if (storedMerchantData && storedMerchantData.length > 0) {
      merchants.value = storedMerchantData
      showEmptyState.value = false
      selectedMerchant.value = merchants.value[0]
      await fetchMerchantProducts(selectedMerchant.value.id)
      loading.value = false
      return
    }

    const currentUser = authStore.currentUser
    if (!currentUser || !currentUser.id) {
      showEmptyState.value = true
      return
    }

    const response = await getMerchants(`?keeper_id=${currentUser.id}`)
    if (response.data) {
      merchants.value = Array.isArray(response.data.merchant_products)
        ? response.data.merchant_products
        : [response.data.merchant_products]
      showEmptyState.value = merchants.value.length === 0
      if (merchants.value.length > 0) {
        selectedMerchant.value = merchants.value[0]
        await fetchMerchantProducts(selectedMerchant.value.id)
      }
    } else {
      merchants.value = []
      showEmptyState.value = true
    }
  } catch (error) {
    console.error('Error fetching merchants:', error)
    merchants.value = []
    showEmptyState.value = true
  } finally {
    loading.value = false
  }
}

const fetchMerchantProducts = async (merchantId) => {
  try {
    loadingProducts.value = true
    const response = await getMerchantProducts(merchantId)
    if (response.data) {
      merchantProducts.value = Array.isArray(response.data.merchant_products)
        ? response.data.merchant_products
        : [response.data.merchant_products]
    } else {
      merchantProducts.value = []
    }
  } catch (error) {
    console.error('Error fetching merchant products:', error)
    merchantProducts.value = []
  } finally {
    loadingProducts.value = false
  }
}

const selectMerchant = async (merchant) => {
  if (!merchant) return
  selectedMerchant.value = merchant
  await fetchMerchantProducts(merchant.id)
}

const refreshProducts = async () => {
  if (selectedMerchant.value) {
    await fetchMerchantProducts(selectedMerchant.value.id)
  }
}

const formatPrice = (price) => new Intl.NumberFormat('id-ID').format(price)

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

onMounted(() => {
  fetchMerchants()
})
</script>
