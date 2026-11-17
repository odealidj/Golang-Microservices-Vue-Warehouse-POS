<template>
  <Layout pageTitle="Manage Products" breadcrumbText="" breadcrumbLink="">
    <section id="Products" class="lte-card-primary flex flex-col flex-1">
      <div id="Header" class="lte-card-header">
        <div>
          <h2 class="lte-card-title">
            <span>All Products</span>
            <span class="lte-badge bg-blue-50 text-lte-primary border border-blue-200">
              {{ totalFilteredCount }} Total
            </span>
          </h2>
          <p class="text-xs text-lte-muted mt-0.5">View and update your product list here.</p>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search product or category..."
            class="lte-input w-56"
          >
          <router-link to="/products/add" class="lte-btn-primary whitespace-nowrap">
            + Add New
          </router-link>
        </div>
      </div>

      <div id="Product-List" class="lte-card-body flex-1 flex flex-col">
        <!-- Loading State -->
        <div v-if="loading" class="flex flex-col flex-1 items-center justify-center py-12 gap-3">
          <div class="animate-spin rounded-full h-9 w-9 border-2 border-lte-primary border-t-transparent"></div>
          <p class="text-sm font-medium text-lte-muted">Memuat data produk...</p>
        </div>

        <!-- Error State -->
        <div v-else-if="error" class="flex flex-col flex-1 items-center justify-center py-12 gap-3 text-center">
          <p class="text-sm font-semibold text-red-600">{{ error }}</p>
          <button @click="retryFetch" class="lte-btn-primary">Coba Lagi</button>
        </div>

        <!-- Products Data Table -->
        <div v-else-if="filteredProducts.length > 0" class="lte-table-wrap">
          <table class="lte-table">
            <thead>
              <tr>
                <th>Product</th>
                <th>Category</th>
                <th>Price</th>
                <th>Popularity</th>
                <th class="text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="product in filteredProducts" :key="product.id">
                <td>
                  <div class="flex items-center gap-3">
                    <img
                      :src="product.thumbnail || galleryDefaultIcon"
                      class="size-11 rounded border border-lte-border object-contain bg-gray-50 shrink-0"
                      alt="product"
                      @error="setDefaultImg"
                    >
                    <span class="font-semibold text-lte-text max-w-[260px] truncate">{{ product.name }}</span>
                  </div>
                </td>
                <td>
                  <div class="flex items-center gap-2">
                    <img
                      v-if="product.category?.photo"
                      :src="product.category.photo"
                      class="size-6 rounded object-contain shrink-0"
                      alt="category"
                      @error="setDefaultImg"
                    >
                    <span class="text-lte-text whitespace-nowrap">{{ product.category?.name || 'Uncategorized' }}</span>
                  </div>
                </td>
                <td class="font-semibold text-lte-primary whitespace-nowrap">
                  Rp {{ formatPrice(product.price) }}
                </td>
                <td>
                  <span
                    class="lte-badge"
                    :class="product.is_popular ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-lte-muted'"
                  >
                    {{ product.is_popular ? 'Popular' : 'Standard' }}
                  </span>
                </td>
                <td class="text-right whitespace-nowrap">
                  <div class="inline-flex items-center gap-1.5">
                    <button @click="viewProduct(product.id)" class="lte-btn-info text-xs py-1 px-2.5">
                      Details
                    </button>
                    <router-link :to="`/products/edit/${product.id}`" class="lte-btn-warning text-xs py-1 px-2.5">
                      Edit
                    </router-link>
                    <button @click="deleteProduct(product.id)" class="lte-btn-danger text-xs py-1 px-2.5">
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Empty State -->
        <div v-else id="Empty-State" class="flex flex-col flex-1 items-center justify-center rounded border border-dashed border-gray-300 gap-3 py-12">
          <img src="@/assets/images/icons/document-text-grey.svg" class="size-10 opacity-70" alt="empty">
          <p class="text-sm font-medium text-lte-muted">Oops, it looks like there's no data yet.</p>
        </div>
      </div>

      <!-- Pagination Footer -->
      <div v-if="totalFilteredCount > 0" class="lte-card-footer flex flex-wrap items-center justify-between gap-3">
        <p class="text-xs text-lte-muted">
          Showing {{ startIndex + 1 }}-{{ endIndex }} of {{ totalFilteredCount }} products
        </p>
        <div class="flex items-center gap-1">
          <button
            @click="previousPage"
            :disabled="currentPage === 1"
            class="lte-btn-default text-xs py-1 px-2.5"
          >
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
          <button
            @click="nextPage"
            :disabled="currentPage === totalPages"
            class="lte-btn-default text-xs py-1 px-2.5"
          >
            Next
          </button>
        </div>
      </div>
    </section>
  </Layout>

  <!-- Product Details Modal -->
  <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <div @click="closeModal" class="absolute inset-0 bg-black/50"></div>
    <div class="relative bg-white rounded border border-lte-border shadow-xl w-full max-w-md overflow-hidden">
      <div class="lte-card-header bg-gray-50">
        <h3 class="lte-card-title">Product Details</h3>
        <button @click="closeModal" class="text-gray-400 hover:text-gray-600 text-xl leading-none">&times;</button>
      </div>
      <div v-if="selectedProduct" class="p-5 space-y-4 max-h-[75vh] overflow-y-auto">
        <div class="flex items-start justify-between gap-4">
          <div class="space-y-1">
            <span class="lte-badge bg-gray-100 text-lte-muted">
              {{ selectedProduct.category?.name || 'Uncategorized' }}
            </span>
            <p class="font-bold text-base text-lte-text">{{ selectedProduct.name }}</p>
            <p class="font-semibold text-sm text-lte-primary">Rp {{ formatPrice(selectedProduct.price) }}</p>
          </div>
          <img
            :src="selectedProduct.thumbnail || galleryDefaultIcon"
            class="size-20 rounded border border-lte-border object-contain bg-gray-50 shrink-0"
            alt="product"
            @error="setDefaultImg"
          >
        </div>
        <div class="border-t border-lte-border pt-3">
          <p class="text-xs font-semibold uppercase text-lte-muted mb-1">Product About</p>
          <p class="text-sm text-lte-text leading-relaxed">{{ selectedProduct.about || 'No description available.' }}</p>
        </div>
        <div class="border-t border-lte-border pt-3 flex items-center justify-between text-sm">
          <span class="text-lte-muted">Popularity</span>
          <span class="font-semibold text-lte-text">{{ selectedProduct.is_popular ? 'Popular' : 'Not Popular' }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import Layout from '@/components/Layout.vue'
import { useProductsStore } from '@/stores/products'
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
import { galleryDefaultIcon } from '@/js/formatters'

export default {
  name: 'ManageProducts',
  components: {
    Layout
  },
  setup() {
    const productsStore = useProductsStore()
    const currentPage = ref(1)
    const itemsPerPage = ref(8)
    const searchQuery = ref('')
    const showModal = ref(false)
    const selectedProduct = ref(null)

    const products = computed(() => productsStore.products)
    const categories = computed(() => productsStore.categories)
    const loading = computed(() => productsStore.loading)
    const error = computed(() => productsStore.error)

    const matchedProducts = computed(() => {
      const q = searchQuery.value.trim().toLowerCase()
      if (!q) return products.value
      return products.value.filter(
        (p) =>
          (p.name || '').toLowerCase().includes(q) ||
          (p.category?.name || '').toLowerCase().includes(q)
      )
    })

    watch(searchQuery, () => {
      currentPage.value = 1
    })

    const totalFilteredCount = computed(() => matchedProducts.value.length)
    const totalPages = computed(() => Math.max(1, Math.ceil(matchedProducts.value.length / itemsPerPage.value)))
    const startIndex = computed(() => (currentPage.value - 1) * itemsPerPage.value)
    const endIndex = computed(() => Math.min(startIndex.value + itemsPerPage.value, matchedProducts.value.length))

    const paginatedProducts = computed(() =>
      matchedProducts.value.slice(startIndex.value, endIndex.value)
    )

    const visiblePages = computed(() => {
      const pages = []
      const maxVisible = 5
      let start = Math.max(1, currentPage.value - Math.floor(maxVisible / 2))
      const end = Math.min(totalPages.value, start + maxVisible - 1)
      if (end - start + 1 < maxVisible) {
        start = Math.max(1, end - maxVisible + 1)
      }
      for (let i = start; i <= end; i++) {
        pages.push(i)
      }
      return pages
    })

    const formatPrice = (price) => new Intl.NumberFormat('id-ID').format(price)

    const goToPage = (page) => {
      currentPage.value = page
    }

    const nextPage = () => {
      if (currentPage.value < totalPages.value) {
        currentPage.value++
      }
    }

    const previousPage = () => {
      if (currentPage.value > 1) {
        currentPage.value--
      }
    }

    const openModal = (product) => {
      selectedProduct.value = product
      showModal.value = true
    }

    const closeModal = () => {
      showModal.value = false
      selectedProduct.value = null
    }

    const viewProduct = (id) => {
      const product = products.value.find((p) => p.id === id)
      if (product) {
        openModal(product)
      }
    }

    const deleteProduct = async (id) => {
      if (confirm('Are you sure you want to delete this product?')) {
        try {
          await productsStore.deleteProduct(id)
          if (paginatedProducts.value.length === 0 && currentPage.value > 1) {
            currentPage.value--
          }
        } catch (error) {
          console.error('Failed to delete product:', error)
          alert(`Gagal menghapus produk: ${error.message}`)
        }
      }
    }

    const fetchProducts = async () => {
      try {
        await productsStore.fetchProducts()
      } catch (error) {
        console.error('Failed to fetch products:', error)
      }
    }

    const retryFetch = async () => {
      await fetchProducts()
    }

    const setDefaultImg = (event) => {
      event.target.src = galleryDefaultIcon
    }

    const handleEscape = (event) => {
      if (event.key === 'Escape' && showModal.value) {
        closeModal()
      }
    }

    onMounted(() => {
      fetchProducts()
      document.addEventListener('keydown', handleEscape)
    })

    onUnmounted(() => {
      document.removeEventListener('keydown', handleEscape)
    })

    return {
      products,
      categories,
      filteredProducts: paginatedProducts,
      totalFilteredCount,
      searchQuery,
      totalPages,
      currentPage,
      startIndex,
      endIndex,
      visiblePages,
      showModal,
      selectedProduct,
      loading,
      error,
      galleryDefaultIcon,
      formatPrice,
      goToPage,
      nextPage,
      previousPage,
      openModal,
      closeModal,
      viewProduct,
      deleteProduct,
      fetchProducts,
      retryFetch,
      setDefaultImg
    }
  }
}
</script>
