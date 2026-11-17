<template>
  <Layout pageTitle="Manage Categories" breadcrumbText="" breadcrumbLink="">
    <main class="flex flex-col flex-1">
      <section id="Categories" class="lte-card-primary flex flex-col flex-1">
        <div id="Header" class="lte-card-header">
          <div>
            <h2 class="lte-card-title">
              <span>All Categories</span>
              <span class="lte-badge bg-blue-50 text-lte-primary border border-blue-200">
                {{ categories.length }} Total
              </span>
            </h2>
            <p class="text-xs text-lte-muted mt-0.5">View and update your Category list here.</p>
          </div>
          <router-link to="/categories/add" class="lte-btn-primary">
            + Add New
          </router-link>
        </div>

        <div id="Category-List" class="lte-card-body flex-1 flex flex-col">
          <!-- Loading State -->
          <div v-if="loading" class="flex flex-col flex-1 items-center justify-center py-12 gap-3">
            <div class="animate-spin rounded-full h-9 w-9 border-2 border-lte-primary border-t-transparent"></div>
            <p class="text-sm font-medium text-lte-muted">Memuat data kategori...</p>
          </div>

          <!-- Error State -->
          <div v-else-if="error" class="flex flex-col flex-1 items-center justify-center py-12 gap-3 text-center">
            <p class="text-sm font-semibold text-red-600">{{ error }}</p>
            <button @click="fetchCategories" class="lte-btn-primary">Coba Lagi</button>
          </div>

          <!-- Empty State -->
          <div v-else-if="categories.length === 0" class="flex flex-col flex-1 items-center justify-center rounded border border-dashed border-gray-300 gap-3 py-12">
            <img src="@/assets/images/icons/document-text-grey.svg" class="size-10 opacity-70" alt="empty">
            <p class="text-sm font-medium text-lte-muted">Oops, it looks like there's no data yet.</p>
          </div>

          <!-- Categories Data Table -->
          <div v-else class="lte-table-wrap">
            <table class="lte-table">
              <thead>
                <tr>
                  <th>Category</th>
                  <th>Tagline / Description</th>
                  <th>Products</th>
                  <th>Status</th>
                  <th class="text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="category in categories" :key="category.id">
                  <td>
                    <div class="flex items-center gap-3">
                      <div class="size-10 rounded-full bg-gray-100 border border-lte-border flex items-center justify-center overflow-hidden shrink-0">
                        <img :src="getCategoryIcon(category.photo || category.icon)" class="size-5 object-contain" alt="icon">
                      </div>
                      <span class="font-semibold text-lte-text whitespace-nowrap">{{ category.name }}</span>
                    </div>
                  </td>
                  <td class="text-lte-muted max-w-[260px] truncate">
                    {{ category.tagline || category.description || '-' }}
                  </td>
                  <td class="font-medium text-lte-text whitespace-nowrap">
                    {{ formatNumber(category.count_product) }} Products
                  </td>
                  <td>
                    <span
                      :class="getCategoryStatus(category).status === 'protected' ? 'bg-yellow-100 text-yellow-800' : 'bg-green-100 text-green-800'"
                      class="lte-badge"
                    >
                      {{ getCategoryStatus(category).status === 'protected' ? 'Protected' : 'Safe' }}
                    </span>
                  </td>
                  <td class="text-right whitespace-nowrap">
                    <div class="inline-flex items-center gap-1.5">
                      <router-link :to="`/categories/edit/${category.id}`" class="lte-btn-warning text-xs py-1 px-2.5">
                        Edit
                      </router-link>
                      <button
                        @click="deleteCategory(category.id, category.name)"
                        :disabled="!getCategoryStatus(category).canDelete"
                        :title="getCategoryStatus(category).message"
                        class="lte-btn-danger text-xs py-1 px-2.5"
                      >
                        Delete
                      </button>
                    </div>
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

<script>
import Layout from '@/components/Layout.vue'
import { getCategories, deleteCategory as deleteCategoryApi } from '@/js/api'
import milkBlackIcon from '@/assets/images/icons/milk-black.svg'
import makeupBlackIcon from '@/assets/images/icons/Makeup-black.svg'
import watchBlackIcon from '@/assets/images/icons/watch-black.svg'
import printerBlackIcon from '@/assets/images/icons/printer-black.svg'
import glassBlackIcon from '@/assets/images/icons/glass-black.svg'
import noteBlackIcon from '@/assets/images/icons/note-2-black.svg'

export default {
  name: 'ManageCategories',
  components: {
    Layout
  },
  data() {
    return {
      categories: [],
      loading: false,
      error: null
    }
  },
  async mounted() {
    await this.fetchCategories()
  },
  methods: {
    async fetchCategories() {
      this.loading = true
      this.error = null
      try {
        const response = await getCategories()
        this.categories = response.data.categories || response
      } catch (error) {
        console.error('Error fetching categories:', error)
        this.error = error.message || 'Gagal mengambil data kategori'
      } finally {
        this.loading = false
      }
    },
    getCategoryIcon(iconName) {
      if (iconName && (iconName.startsWith('http://') || iconName.startsWith('https://'))) {
        return iconName
      }
      const iconMap = {
        'milk-black': milkBlackIcon,
        'Makeup-black': makeupBlackIcon,
        'watch-black': watchBlackIcon,
        'printer-black': printerBlackIcon,
        'glass-black': glassBlackIcon
      }
      return iconMap[iconName] || noteBlackIcon
    },
    formatNumber(num) {
      return new Intl.NumberFormat('en-US').format(num || 0)
    },
    getCategoryStatus(category) {
      if (category.count_product > 0) {
        return {
          status: 'protected',
          message: `Protected - ${this.formatNumber(category.count_product)} products`,
          canDelete: false
        }
      }
      return {
        status: 'deletable',
        message: 'No products - Safe to delete',
        canDelete: true
      }
    },
    async deleteCategory(categoryId, categoryName) {
      const category = this.categories.find((cat) => cat.id === categoryId)
      if (category && category.count_product > 0) {
        alert(`Cannot delete category "${categoryName}" because it has ${this.formatNumber(category.count_product)} products. Please remove all products first.`)
        return
      }
      if (confirm(`Apakah Anda yakin ingin menghapus kategori "${categoryName}"?`)) {
        try {
          await deleteCategoryApi(categoryId)
          this.categories = this.categories.filter((cat) => cat.id !== categoryId)
          alert(`Kategori "${categoryName}" berhasil dihapus!`)
        } catch (error) {
          console.error('Error deleting category:', error)
          alert(`Gagal menghapus kategori: ${error.message}`)
        }
      }
    }
  }
}
</script>
