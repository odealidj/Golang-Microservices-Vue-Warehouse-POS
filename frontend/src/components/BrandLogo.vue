<template>
  <div
    class="inline-flex items-center shrink-0 select-none"
    :class="size === 'lg' ? 'gap-3' : 'gap-2.5 h-8'"
    aria-label="CODINA logo"
  >
    <img
      src="@/assets/images/logos/logo-notext.svg"
      :class="size === 'lg' ? 'h-9 w-auto' : 'h-7 w-auto'"
      alt="CODINA mark"
    >
    <span
      class="font-extrabold tracking-[0.08em] text-[#292D32] leading-none"
      :class="size === 'lg' ? 'text-[28px]' : 'text-[22px]'"
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
  }
})
</script>
