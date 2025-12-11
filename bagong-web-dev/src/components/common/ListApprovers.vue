<template>
  <div class="min-w-[200px] overflow-auto">
    <table>
      <tr v-for="(approvers, idx) in props.approvers"
        :key="idx"
        >
        <td class="text-[0.8rem]">{{ approvers }}</td>
      </tr>
    </table>
  </div>
</template>
<script setup>
const props = defineProps({
  approvers: { type: Array, default: () => [] },
});
</script>
