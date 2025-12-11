<template>
  <div>
    <label class="input_label">{{ title }}</label>
    <data-list
      ref="listControl"
      hide-title
      no-gap
      grid-editor
      grid-hide-search
      grid-hide-sort
      grid-hide-refresh
      grid-hide-detail
      grid-hide-select
      grid-no-confirm-delete
      :init-app-mode="data.appMode"
      :grid-mode="data.appMode"
      new-record-type="grid"
      :grid-fields="['TrayekID']"
      grid-config="/bagong/corridor/gridconfig"
      form-config="/bagong/corridor/formconfig"
      grid-auto-commit-line
      @grid-row-add="newRecord"
      @grid-row-delete="onGridRowDelete"
      @gridRefreshed="onGridRefreshed"
    >
      <template #grid_TrayekID="{ item }">
        <s-input
          class="min-w-[100px]"
          hide-label
          use-list
          v-model="item.TrayekID"
          lookup-url="/bagong/trayek/find"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
        />
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted, watch } from "vue";
import { DataList, SInput } from "suimjs";

const props = defineProps({
  title: { type: String, default: "" },
  modelValue: { type: Array, default: () => [] },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const listControl = ref(null);

const data = reactive({
  appMode: "grid",
  records:
    props.modelValue == null || props.modelValue == undefined
      ? []
      : props.modelValue,
});

function newRecord(r) {
  r.Name = "";
  r.TrayekID = "";

  data.records.push(r);
  updateItems();
}

function onGridRowDelete(_, index) {
  const newRecords = data.records.filter((_, idx) => {
    return idx != index;
  });
  data.records = newRecords;
  listControl.value.setGridRecords(data.records);
  updateItems();
}

function updateItems() {
  listControl.value.setGridRecords(data.records);
  emit("update:modelValue", data.records);
}
function onGridRefreshed(){
  setTimeout(() => {
    updateItems();
  }, 500);
}
 

watch(
  () => data.records,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
</script>
