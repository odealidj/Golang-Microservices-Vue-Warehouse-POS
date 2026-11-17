<template>
  <AdminLteTopBar
    :title="title"
    :user-name="userName"
    :user-role="userRole"
    :breadcrumb-text="breadcrumbText"
    :breadcrumb-link="breadcrumbLink"
  />
</template>

<script setup>
import AdminLteTopBar from '@/components/AdminLteTopBar.vue'

defineProps({
  title: {
    type: String,
    default: 'Merchant Overview'
  },
  userName: {
    type: String,
    default: 'Raratih Aqilla'
  },
  userRole: {
    type: String,
    default: 'Keeper'
  },
  breadcrumbText: {
    type: String,
    default: null
  },
  breadcrumbLink: {
    type: String,
    default: '/transactions'
  }
})
</script>
