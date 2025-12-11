<template>
  <div class="flex flex-col gap-2">
    <s-grid
      v-show="data.typeMoveIn == 'line'"
      ref="listControl"
      class="w-full grid-line-items"
      :editor="data.value.Status == '' || data.value.Status == 'Draft'"
      hide-search
      hide-select
      hide-sort
      hide-new-button
      hide-delete-button
      hide-refresh-button
      :hide-detail="true"
      hide-paging
      hide-action
      auto-commit-line
      no-confirm-delete
      :config="data.gridCfg"
      form-keep-label
    >
      <template #item_ItemID="{ item, idx }">
        {{ item.ItemName }}
      </template>
      <template #item_UoM="{ item }">
        {{ item.UnitName }}
      </template>
      <template #item_Aisle="{ item }">
        {{ item.AisleName }}
      </template>
      <template #item_Section="{ item }">
        {{ item.SectionName }}
      </template>
      <template #item_Box="{ item }">
        {{ item.BoxName }}
      </template>
      <template #item_Remarks="{ item }">
        <div
          class="text-center"
          :class="
            item.Remarks == 'MINUS'
              ? 'text-minus'
              : item.Remarks == 'OVER'
              ? 'text-over'
              : 'text-ok'
          "
        >
          {{ item.Remarks }}
        </div>
      </template>
      <template
        v-if="['NeedToReview'].includes(props.item.Status)"
        #item_NoteAdjustment="{ item }"
      >
        <s-input
          ref="refNoteAdjustment"
          v-model="item.NoteAdjustment"
          :disabled="!['NeedToReview'].includes(props.item.Status)"
          multiRow="2"
          :keepErrorSection="true"
        ></s-input>
      </template>
    </s-grid>
  </div>
</template>
<script setup>
import { onMounted, inject, computed, watch, reactive, ref } from "vue";
import { loadGridConfig, util, SInput, SGrid } from "suimjs";
const axios = inject("axios");
const refItemID = ref(null);
const refSKU = ref(null);
const refUom = ref(null);
const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  item: { type: Object, default: () => {} },
  itemID: { type: String, default: () => "" },
  gridConfig: { type: String, default: () => "" },
  gridRead: { type: String, default: () => "" },
  formMode: { type: String, default: () => "new" },
  activeFields: { type: Array, default: () => [] },
  readOnly: { type: Boolean, defaule: false },
  hideDetail: { type: Boolean, defaule: false },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});

const listControl = ref(null);

const data = reactive({
  value: props.modelValue,
  typeMoveIn: "line",
  fromCfgBatch: {},
  fromCfgSN: {},
  gridCfg: {},
});

function getDataValue() {
  return listControl.value.getRecords();
}
onMounted(() => {
  loadGridConfig(axios, props.gridConfig).then(
    (r) => {
      let tbLine = [
        "ItemID",
        "UoM",
        "Aisle",
        "Section",
        "Box",
        "QtyActual",
        "QtyInSystem",
        "Gap",
        "Remarks",
        "NoteStaff",
        "Note",
        "NoteAdjustment",
      ];
      let InventoryDimension = [
        {
          field: "Aisle",
          kind: "Text",
          label: "Aisle",
          readType: "show",
          input: {
            field: "Aisle",
            label: "Aisle",
            hint: "",
            hide: false,
            placeHolder: "Aisle",
            kind: "text",
            disable: false,
            required: false,
            multiple: false,
          },
        },
        {
          field: "Section",
          kind: "Text",
          label: "Section",
          readType: "show",
          input: {
            field: "Section",
            label: "Section",
            hint: "",
            hide: false,
            placeHolder: "Section",
            kind: "text",
            disable: false,
            required: false,
            multiple: false,
          },
        },
        {
          field: "Box",
          kind: "Text",
          label: "Box",
          readType: "show",
          input: {
            field: "Box",
            label: "Box",
            hint: "",
            hide: false,
            placeHolder: "Box",
            kind: "text",
            disable: false,
            required: false,
            multiple: false,
          },
        },
      ];
      const _fields = [...r.fields, ...InventoryDimension].filter((o) => {
        if (["UoM", "QtyInSystem", "QtyActual", "Gap"].includes(o.field)) {
          o.width = "300px";
        } else if (
          [
            "Description",
            "Remarks",
            "NoteStaff",
            "Note",
            "NoteAdjustment",
          ].includes(o.field)
        ) {
          o.width = "800px";
        } else {
          o.width = "400px";
        }
        o.idx = tbLine.indexOf(o.field);
        return tbLine.includes(o.field);
      });
      _fields.map((o) => {
        if (o.field == "ItemID") {
          o.label = "Item Varian";
        }
      });
      data.gridCfg = {
        ...r,
        fields: _fields.sort((a, b) => (a.idx > b.idx ? 1 : -1)),
      };
    },
    (e) => util.showError(e)
  );
  axios
    .post(props.gridRead, {
      Skip: 0,
      Take: 0,
      Sort: ["_id"],
    })
    .then(
      (r) => {
        if (listControl.value) {
          r.data.data.map(function (v) {
            if (v.InventoryDimension) {
              v.AisleID = v.InventoryDimension.AisleID;
              v.BoxID = v.InventoryDimension.BoxID;
              v.SectionID = v.InventoryDimension.SectionID;
            }
            return v;
          });
          listControl.value.setRecords(r.data.data);
        }
      },
      (e) => util.showError(e)
    );
});
defineExpose({
  getDataValue,
});
</script>
<style>
.tb-line > div:nth-child(2) > div {
  overflow-x: auto;
  padding-bottom: 100px;
}
.tb-line > div:nth-child(2) > div > table {
  width: calc(100% + 40%) !important;
}
.text-minus {
  color: red;
  font-weight: 600;
}

.text-ok {
  color: black;
  font-weight: 600;
}

.text-over {
  color: green;
  font-weight: 600;
}
</style>
