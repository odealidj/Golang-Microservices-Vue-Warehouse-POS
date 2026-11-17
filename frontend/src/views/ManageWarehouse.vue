<template>
  <Layout pageTitle="Manage Warehouses" breadcrumbText="" breadcrumbLink="">
    <main class="flex flex-col flex-1">
      <section id="Warehouses" class="lte-card-primary flex flex-col flex-1">
        <div id="Header" class="lte-card-header">
          <div>
            <h2 class="lte-card-title">
              <span>All Warehouses</span>
              <span class="lte-badge bg-blue-50 text-lte-primary border border-blue-200">
                {{ totalRecords }} Total
              </span>
            </h2>
            <p class="text-xs text-lte-muted mt-0.5">View and update your Warehouses list here.</p>
          </div>
          <router-link to="/warehouses/add" class="lte-btn-primary">
            + Add New
          </router-link>
        </div>

        <div id="Warehouse-List" class="lte-card-body flex-1 flex flex-col">
          <!-- Empty State -->
          <div v-if="warehouses.length === 0" class="flex flex-col flex-1 items-center justify-center rounded border border-dashed border-gray-300 gap-3 py-12">
            <img src="@/assets/images/icons/document-text-grey.svg" class="size-10 opacity-70" alt="empty">
            <p class="text-sm font-medium text-lte-muted">Oops, it looks like there's no data yet.</p>
          </div>

          <!-- Warehouses Data Table -->
          <div v-else class="lte-table-wrap">
            <table class="lte-table">
              <thead>
                <tr>
                  <th>Warehouse</th>
                  <th>Phone</th>
                  <th>Stocked Products</th>
                  <th class="text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="warehouse in warehouses" :key="warehouse.id">
                  <td>
                    <div class="flex items-center gap-3">
                      <img
                        :src="getWarehouseImage(warehouse.photo)"
                        class="size-11 rounded border border-lte-border object-contain bg-gray-50 shrink-0"
                        alt="warehouse"
                      >
                      <span class="font-semibold text-lte-text whitespace-nowrap">{{ warehouse.name }}</span>
                    </div>
                  </td>
                  <td class="text-lte-muted whitespace-nowrap">{{ warehouse.phone }}</td>
                  <td class="whitespace-nowrap">
                    <span class="lte-badge bg-gray-100 text-lte-text">
                      {{ formatNumber(warehouse.count_product) }} Products
                    </span>
                  </td>
                  <td class="text-right whitespace-nowrap">
                    <div class="inline-flex items-center gap-1.5">
                      <router-link :to="`/warehouses/${warehouse.id}`" class="lte-btn-info text-xs py-1 px-2.5">
                        Details
                      </router-link>
                      <router-link :to="`/warehouses/edit/${warehouse.id}`" class="lte-btn-warning text-xs py-1 px-2.5">
                        Edit
                      </router-link>
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
import { getWarehouses } from '@/js/api'
import buildingsIcon from '@/assets/images/icons/buildings-2-black.svg'

export default {
  name: 'ManageWarehouse',
  components: {
    Layout
  },
  data() {
    return {
      warehouses: [],
      totalRecords: 0
    }
  },
  async mounted() {
    await this.fetchWarehouses()
  },
  methods: {
    async fetchWarehouses() {
      try {
        const response = await getWarehouses()
        this.warehouses = response.data.warehouses || []
        this.totalRecords = response.data.pagination.total_records || 0
      } catch (err) {
        console.error('Error fetching warehouses:', err)
      }
    },
    getWarehouseImage(photo) {
      return photo || buildingsIcon
    },
    formatNumber(num) {
      return new Intl.NumberFormat('en-US').format(num || 0)
    }
  }
}
</script>
