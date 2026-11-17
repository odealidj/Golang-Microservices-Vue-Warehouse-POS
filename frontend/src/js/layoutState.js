import { ref } from 'vue'

export const sidebarCollapsed = ref(false)
export const mobileSidebarOpen = ref(false)

/**
 * Toggle desktop mini-sidebar or mobile drawer depending on viewport width.
 */
export function toggleSidebar() {
  if (typeof window !== 'undefined' && window.innerWidth < 1024) {
    mobileSidebarOpen.value = !mobileSidebarOpen.value
    return
  }
  sidebarCollapsed.value = !sidebarCollapsed.value
}

export function closeMobileSidebar() {
  mobileSidebarOpen.value = false
}
