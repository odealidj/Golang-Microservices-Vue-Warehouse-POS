<template>
  <div id="main-container" class="flex flex-1 bg-lte-bg">
    <!-- AdminLTE Sidebar (Manager) -->
    <AdminLteSidebar role-mode="manager" />

    <!-- Main Content Wrapper -->
    <div id="Content" class="flex flex-col flex-1 p-6 pt-0 bg-lte-bg min-w-0">
      <AdminLteTopBar
        :title="pageTitle"
        :breadcrumb-text="breadcrumbText"
        :breadcrumb-link="breadcrumbLink"
      />

      <!-- Page Content -->
      <div class="flex-1 flex flex-col">
        <slot></slot>
      </div>

      <!-- AdminLTE Main Footer -->
      <footer class="mt-8 -mx-6 px-6 py-3 bg-white border-t border-lte-border text-xs text-lte-muted flex flex-col sm:flex-row items-center justify-between gap-2">
        <span><strong>CODINA</strong> &mdash; Warehouse &amp; Retail POS Platform</span>
        <span>AdminLTE-style Shell</span>
      </footer>
    </div>
  </div>
</template>

<script setup>
import AdminLteSidebar from '@/components/AdminLteSidebar.vue'
import AdminLteTopBar from '@/components/AdminLteTopBar.vue'

defineProps({
  pageTitle: {
    type: String,
    default: 'Dashboard'
  },
  breadcrumbText: {
    type: String,
    default: null
  },
  breadcrumbLink: {
    type: String,
    default: '/products'
  }
})
</script>
