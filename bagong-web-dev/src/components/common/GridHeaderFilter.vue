<template>
  <div class="flex flex-1 flex-wrap gap-3 justify-start grid-header-filter">
    <slot name="filter_1" :item="data.item"> </slot>
    <slot name="filters" :item="data.item">
      <s-input
        v-if="!hideFilterText && !hideAll"
        class="w-[200px] filter-text"
        :label="builtTextLabel()"
        v-model="data.item.Text"
      />
      <div class="flex gap-1">
        <s-input
          class="filter-date-from"
          v-if="!hideFilterDate && !hideAll"
          label="Date From "
          kind="date"
          v-model="data.item.DateFrom"
        />
        <s-input
          class="filter-date-to"
          v-if="!hideFilterDate && !hideAll"
          label="Date To"
          kind="date"
          v-model="data.item.DateTo"
        />
      </div>
      <s-input
        v-if="!hideFilterStatus && !hideAll"
        class="min-w-[200px] filter-status"
        label="Status"
        use-list
        multiple
        :items="['DRAFT', 'SUBMITTED', 'READY', 'REJECTED', 'POSTED']"
        v-model="data.item.Status"
      />
    </slot>
    <slot name="filter_2" :item="data.item"> </slot>
  </div>
</template>
<script setup>
import { reactive, ref, watch, onMounted, inject } from "vue";
import { SInput, util } from "suimjs";
import helper from "@/scripts/helper.js";
const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  autoRefresh: { type: Boolean, default: false },
  hideAll: { type: Boolean, default: false },
  hideFilterText: { type: Boolean, default: false },
  hideFilterDate: { type: Boolean, default: false },
  hideFilterStatus: { type: Boolean, default: false },
  fieldsText: { type: Array, default: ["Text"] },
  customTextLabel: {tyop: String, default: ""},
  fieldsDate: { type: Array, default: ["TrxDate"] },
  fieldsStatus: { type: Array, default: ["Status"] },
  operator: { type: String, default: "$and" },
  filterTextWithId: { type: Boolean, default: false },
});
const emit = defineEmits({
  "update:modelValue": null,
  preChange: null,
  change: null,
  initNewItem: null,
});
const data = reactive({
  filter:
    props.modelValue == null || props.modelValue == undefined
      ? null
      : props.modelValue,
  disableWatch: true,
  item: {
    Text: "",
    DateFrom: null,
    DateTo: null,
    Status: [],
  },
});
watch(
  () => data.filter,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);

const debounce = createDebounce();
watch(
  () => JSON.stringify(data.item),
  async (v) => {
    if (data.disableWatch === true) return;
    const Filter = JSON.parse(v);
    const filters = [];

    if (!props.hideAll && !props.hideFilterText && Filter.Text && Filter.Text != "") {
      props.fieldsText.forEach((e) => {
        if (props.filterTextWithId) {
          filters.push({
            Op: "$or",
            Items: [
              {
                Op: "$contains",
                Field: e,
                Value: [Filter.Text],
              },
              {
                Op: "$contains",
                Field: "_id",
                Value: [Filter.Text],
              },
            ],
          });
        } else {
          filters.push({
            Op: "$contains",
            Field: e,
            Value: [Filter.Text],
          });
        }
      });
    }

    if (!props.hideAll && !props.hideFilterDate && Filter.DateFrom != null) {
      props.fieldsDate.forEach((f) => {
        filters.push({
          Op: "$gte",
          Field: f,
          Value: helper.formatFilterDate(Filter.DateFrom),
        });
      });
    }

    if (!props.hideAll && !props.hideFilterDate && Filter.DateTo != null) {
      props.fieldsDate.forEach((f) => {
        filters.push({
          Op: "$lte",
          Field: f,
          Value: helper.formatFilterDate(Filter.DateTo, true),
        });
      });
    }

    if (!props.hideAll && !props.hideFilterStatus && Filter.Status.length > 0) {
      props.fieldsStatus.forEach((f) => {
        filters.push({
          Op: "$in",
          Field: f,
          Value: [...Filter.Status],
        });
      });
    }

    emit("preChange", Filter, filters);

    if (filters.length > 0) {
      data.filter = {
        Op: props.operator,
        Items: filters,
      };
    } else {
      data.filter = undefined;
    }

    debounce(() => {
      emit("change");
    }, 500);
  }
);

function createDebounce() {
  let timeout = null;
  return function (fnc, delayMs) {
    clearTimeout(timeout);
    timeout = setTimeout(() => {
      fnc();
    }, delayMs || 500);
  };
}
function reset() {
  data.disableWatch = true;
  data.item = {
    Text: "",
    DateFrom: null,
    DateTo: null,
    Status: [],
  };
  emit("initNewItem", data.item);
  data.filter = null;
  data.disableWatch = false;
}
function init() {
  data.disableWatch = true;
  emit("initNewItem", data.item);

  util.nextTickN(2, () => {
    data.disableWatch = false;
  });
}
function builtTextLabel() {
  return props.customTextLabel !== "" ? props.customTextLabel : props.fieldsText.join(', ')
}
onMounted(() => {
  init();
});
defineExpose({
  reset,
});
</script>
