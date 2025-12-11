<template>
  <div class="flex flex-col gap-2">
    <s-grid
      v-model="data.listSpec"
      ref="listControl"
      class="w-full tb-line grid-line-items"
      editor
      hide-search
      hide-select
      hide-sort
      :hide-new-button="false"
      :hide-delete-button="false"
      :hide-detail="true"
      hide-refresh-button
      auto-commit-line
      no-confirm-delete
      :config="data.gridCfg"
      form-keep-label
      :hide-paging="true"
      @new-data="newRecord"
      @delete-data="onDelete"
    >
      <template #item_IsActive="{ item, idx }">
        <div class="row_action">
          <s-input
            v-model="item.IsActive"
            hide-label
            kind="checkbox"
            class="w-full flex justify-center"
          ></s-input>
        </div>
      </template>
    </s-grid>
  </div>
</template>
<script setup>
import { onMounted, watch, inject, reactive, ref } from "vue";
import { loadGridConfig, util, DataList, SGrid, SInput } from "suimjs";
const axios = inject("axios");
const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  item: { type: Object, default: () => {} },
  typeSpec: { type: String, default: () => "" },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});

const listControl = ref(null);
const data = reactive({
  appMode: "grid",
  formMode: "edit",
  gridCfg: {
    fields: [],
    setting: {},
  },
  listSpec: [],
});

function newRecord() {
  const record = {};
  record.ItemID = props.item._id;
  record.SKU = "";
  record.OtherName = "";
  record.SpecVariantID = "";
  record.SpecSizeID = "";
  record.SpecGradeID = "";
  record.IsActive = true;
  listControl.value.setRecords([record, ...listControl.value.getRecords()]);
  updateRecords();
}

function onDelete(record, index) {
  const newRecords = listControl.value.getRecords().filter((dt, idx) => {
    return idx != index;
  });
  listControl.value.setRecords(newRecords);
  updateRecords();
}

function getDataValue() {
  return listControl.value.getRecords();
}

function updateRecords() {
  data.listSpec = listControl.value.getRecords();
}

watch(
  () => props.item.PhysicalDimension,
  (nv) => {
    let tbLine = ["SKU", "OtherName"];
    if (nv.IsEnabledSpecVariant) {
      tbLine.push("SpecVariantID");
    }
    if (nv.IsEnabledSpecGrade) {
      tbLine.push("SpecGradeID");
    }
    if (nv.IsEnabledSpecSize) {
      tbLine.push("SpecSizeID");
    }
    tbLine.push("IsActive");
    util.nextTickN(2, () => {
      data.gridCfg.fields.map((o) => {
        if (tbLine.includes(o.field)) {
          o.readType = "show";
          o.input.hide = false;
        } else {
          o.readType = "hide";
          o.input.hide = true;
        }
        o.idx = tbLine.indexOf(o.field);
        return o;
      });
      data.gridCfg = {
        ...data.gridCfg,
        fields: data.gridCfg.fields.sort((a, b) => (a.idx > b.idx ? 1 : -1)),
      };
    });
  },
  { deep: true }
);

function getsSpec() {
  axios
    .post(`/tenant/itemspec/gets?ItemID=${props.item._id}`, {
      Skip: 0,
      Take: 0,
      Sort: ["_id"],
    })
    .then(
      (r) => {
        data.listSpec = r.data.data;
      },
      (e) => util.showError(e)
    );
}

onMounted(() => {
  loadGridConfig(axios, `/tenant/itemspec/gridconfig`).then(
    (r) => {
      let tbLine = ["SKU", "OtherName"];
      if (props.item.PhysicalDimension.IsEnabledSpecVariant) {
        tbLine.push("SpecVariantID");
      }
      if (props.item.PhysicalDimension.IsEnabledSpecGrade) {
        tbLine.push("SpecGradeID");
      }
      if (props.item.PhysicalDimension.IsEnabledSpecSize) {
        tbLine.push("SpecSizeID");
      }
      tbLine.push("IsActive");
      r.fields.map((o) => {
        if (tbLine.includes(o.field)) {
          o.readType = "show";
          o.input.hide = false;
        } else {
          o.readType = "hide";
          o.input.hide = true;
        }
        o.idx = tbLine.indexOf(o.field);
        return o;
      });

      data.gridCfg = {
        ...r,
        fields: r.fields.sort((a, b) => (a.idx > b.idx ? 1 : -1)),
      };
      getsSpec();
    },
    (e) => util.showError(e)
  );
});
defineExpose({
  getDataValue,
});
</script>
