<template>
  <div
    class="p-1 my-2 rounded-[0.5rem] inline-block font-semibold text-[0.7rem] status-text"
    :class="[
      moduleId.toLowerCase().replace(' ', '_'),
      txt.toLowerCase().replace(' ', '_'),
    ]"
  >
    {{ txt }}
  </div>
</template>
<script setup>
const props = defineProps({
  txt: { type: Number, default: "" },
  moduleId: { type: String, default: "" },
});
</script>
<style scoped>
.draft {
  @apply text-[#444444] bg-[#F8F8F9];
}
.pending {
  @apply text-[#795410] bg-[#f8f5ec];
}
.approved {
  @apply text-[#397341] bg-[#F2FCF8];
}
.submitted {
  @apply text-[#487A99] bg-[#EFF9FF];
}
.posted {
  @apply text-[#397341] bg-[#F2FCF8];
}
.rejected {
  @apply text-[#922C2D] bg-[#FFF7F6];
}
.ready {
  @apply text-[#767a0a] bg-[#f3f3ec];
}
.open {
  @apply text-[#922C2D] bg-[#FFF7F6];
}
.close {
  @apply text-[#397341] bg-[#F2FCF8];
}
.passed {
  @apply text-[#397341] bg-[#F2FCF8];
}
.failed {
  @apply text-[#922C2D] bg-[#FFF7F6];
}
.canceled {
  @apply text-[#922C2D] bg-[#FFF7F6];
}
.not_selected {
  @apply text-[#060606] bg-[#e6e5e4];
}
.expired {
  @apply text-[#922C2D] bg-[#FFF7F6];
}
.active {
  @apply text-[#397341] bg-[#F2FCF8];
}
</style>
