<template>
  <div :class="`${prefixClass} suim_form`">
    <div :class="`${prefixClass} mb-2 flex header`">
      <div :class="`${prefixClass} flex tab_container grow`" v-if="tabs.length > 1">
        <div
          v-for="(tabTitle, tabIdx) in tabs"
          @click="
            data.currentTab = tabIdx;
            emit('activeTab', tabs[data.currentTab]);
          "
          :class="{
            tab_selected: data.currentTab == tabIdx,
            tab: data.currentTab != tabIdx,
          }"
          :key="tabIdx"
        >
          {{ tabTitle }}
        </div>
      </div>
    </div>
    <div v-for="(tabName, tabIdx) in tabs" :key="tabIdx">
      <div v-show="data.currentTab == tabIdx">
        <slot :name="'tab_' + tabName.replaceAll(' ', '_') + '_body'"></slot>
      </div>
    </div>
  </div>
</template>
<script setup>
import { reactive } from "vue";
const props = defineProps({
  tabs: { type: Array, default: [] },
  tabTextLeft: { type: Boolean, default: false },
  prefixClass: { type: String, default: '' },
});
const emit = defineEmits({
  activeTab: null,
});
const data = reactive({
  currentTab: 0,
});

function setCurrentTab(tabidx) {
  data.currentTab = tabidx;
}

defineExpose({
  setCurrentTab,
});
</script>
