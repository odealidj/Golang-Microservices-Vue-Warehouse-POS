<template>
  <div
    class="inline-flex items-center shrink-0 select-none"
    :class="compact ? 'justify-center' : (size === 'lg' ? 'gap-3' : 'gap-2.5 h-8')"
    aria-label="CODINA logo"
  >
    <img
      src="@/assets/images/logos/logo-notext.svg"
      :class="size === 'lg' ? 'h-9 w-auto' : 'h-7 w-auto'"
      alt="CODINA mark"
    >
    <span
      v-if="!compact"
      class="font-extrabold tracking-[0.08em] leading-none"
      :class="[
        size === 'lg' ? 'text-[28px]' : 'text-[20px]',
        variant === 'light' ? 'text-white' : 'text-[#292D32]'
      ]"
    >
      CODINA
    </span>
  </div>
</template>

<script setup>
defineProps({
  size: {
    type: String,
    default: 'md'
  },
  variant: {
    type: String,
    default: 'dark'
  },
  compact: {
    type: Boolean,
    default: false
  }
})
</script>
