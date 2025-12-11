<template>
  <div class="grid gap-1" :class="[hideStatus ? 'grid2' : 'grid3']">
    <label class="input_label" v-if="!hideStatus"> Status </label>
    <label class="input_label"> Description </label>
    <label class="input_label"> Action </label>
    <div class="" v-if="!hideStatus">
      <status-text :txt="pica.Status" />
    </div>
    <div class="">{{ formatDesc(pica.FindingDescription) }}</div>
    <div class="">{{ formatDesc(pica.Comment) }}</div>
  </div>
</template>
<script setup>
import StatusText from "@/components/common/StatusText.vue";
const props = defineProps({
  pica: { type: Object, default: {} },
  hideStatus: { type: Boolean, default: false },
});

function formatDesc(params) {
  return params ? params : "-";
}
</script>
<style scoped>
.grid3 {
  @apply grid-cols-3;
}
.grid2 {
  @apply grid-cols-2;
}
</style>
