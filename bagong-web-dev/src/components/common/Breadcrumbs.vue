<template>
  <div>
    <ul class="flex flex-row content-center gap-1 text-xs mb-4">
      <template v-for="(item, index) in items" :key="index">
        <li>
          <slot name="title" :item="item">
            <div class="flex items-center gap-1">
              <router-link
                v-if="item.url != '' && items.length < index + 1"
                :to="item.url"
                class="text-primary"
              >
                 {{ item.label }}</router-link
              >
              <template v-else>
                {{ item.label }}
              </template>
            </div>
          </slot>
        </li>
        <li v-if="index + 1 < items.length">
          <slot name="divider">
            <slot> {{ divider }}</slot>
          </slot>
        </li>
      </template>
    </ul>
  </div>
</template>
<script setup>
const props = defineProps({
  items: { type: Array, default: [] },
  divider: { type: String, default: "/" },
});
</script>
<style scoped>
ul li:last-child{
  @apply opacity-50;
}
</style>