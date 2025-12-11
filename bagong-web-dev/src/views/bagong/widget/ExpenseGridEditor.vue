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
      :grid-fields="['_id']"
      grid-config="/tenant/expensetype/gridconfig"
      form-config="/tenant/expensetype/formconfig"
      grid-auto-commit-line
      @grid-row-field-changed="gridRowFieldChanged"
      @grid-row-add="newRecord"
      @grid-row-delete="onGridRowDelete"
    >
      <template #grid__id="{ item }">
        <s-input
          class="min-w-[100px]"
          hide-label
          use-list
          v-model="item._id"
          lookup-url="/tenant/expensetype/find"
          :lookup-payload-builder="
            data.groupId.length > 0 ? expensePayload : undefined
          "
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          @change="handleChange"
        />
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted, inject, watch } from "vue";
import { DataList, SInput, loadGridConfig, util } from "suimjs";

const props = defineProps({
  title: { type: String, default: "" },
  modelValue: { type: Array, default: () => [] },
  page: { type: String, default: "" },
  groupIdValue: { type: Array, default: () => [] },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const listControl = ref(null);
const axios = inject("axios");

const data = reactive({
  appMode: "grid",
  records:
    props.modelValue == null || props.modelValue == undefined
      ? []
      : props.modelValue,
  groupId: props.groupIdValue,
});

function newRecord(r) {
  r._id = "";
  r.Name = "";
  r.ExpenseType = "";
  r.Value = "";

  data.records.push(r);
  updateItems();
}

function handleChange(_, id) {
  setTimeout(() => {
    const records = data.records;
    const matchingIndices = records.reduce((indices, record, index) => {
      if (record._id === id) {
        indices.push(index);
      }
      return indices;
    }, []);
    if (matchingIndices.length > 0) {
      axios.post(`/tenant/expensetype/find?_id=${id}`).then((res) => {
        const record = res.data[0];
        matchingIndices.forEach((index) => {
          records[index] = {
            ...record,
            Value: record.Value,
          };
        });
        updateItems();
      });
    }
  }, 500);
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

function expensePayload() {
  return {
    Take: 20,
    Sort: ["Name"],
    Select: ["Name", "_id"],
    Where: {
      Op: "$contains",
      Field: "GroupID",
      Value: data.groupId,
    },
  };
}

function gridRowFieldChanged(name, v1, v2, old, record) {
  if (name == "GroupID") {
    data.groupId = typeof v1 == "string" ? [v1] : [];
  }
}

onMounted(() => {
  loadGridConfig(axios, "/tenant/expensetype/gridconfig").then(
    (r) => {
      util.nextTickN(2, () => {
        if (props.page != "Trayek") {
          listControl.value.removeGridField("ExpenseType");
        }
        if (props.page == "Trayek") {
          listControl.value.removeGridField("LedgerAccountID");
          listControl.value.removeGridField("GroupID");
        }
        listControl.value.removeGridField("IsActive");
        setTimeout(() => {
          updateItems();
        }, 500);
      });
    },
    (e) => util.showError(e)
  );
});

watch(
  () => data.records,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
</script>
