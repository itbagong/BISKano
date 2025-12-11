<template>
  <div class="flex flex-col gap-2">
    <data-list
      ref="listControl"
      title="Posting Profile"
      hide-title
      no-gap
      :grid-hide-detail="gridHideDetail || readOnly"
      :grid-editor="!readOnly"
      :grid-hide-delete="readOnly"
      :grid-hide-control="readOnly"
      grid-hide-select
      grid-hide-search
      grid-hide-sort
      grid-hide-refresh
      grid-no-confirm-delete
      gridAutoCommitLine
      init-app-mode="grid"
      grid-mode="grid"
      form-keep-label
      new-record-type="grid"
      :grid-config="props.gridConfigUrl"
      :form-config="props.formConfigUrl"
      :form-fields="[
        'TagObjectID1',
        'TagObjectID2',
        'OffsetAccount',
        'Dimension',
        'Account',
        'PaymentType',
        'ChequeGiroID',
        'Ignore',
      ]"
      :grid-fields="[
        'TagObjectID1',
        'TagObjectID2',
        'OffsetAccount',
        'Dimension',
        'Account',
        'PaymentType',
        'ChequeGiroID',
        'Ignore',
      ]"
      :init-form-mode="data.formMode"
      @grid-row-add="onGridNewRecord"
      @form-field-change="onFormFieldChanged"
      @form-edit-data="openForm"
      @alter-grid-config="onAlterGridConfig"
      @grid-row-delete="onGridRowDelete"
      @grid-row-field-changed="onGridRowFieldChanged"
      @grid-refreshed="gridRefreshed"
      @grid-row-save="onGridRowSave"
      @post-save="onFormPostSave"
      formHideCancel
      :formHideSubmit="readOnly"
      :form-tabs-edit="data.formTabs"
      :form-tabs-new="data.formTabs"
      :form-tabs-view="data.formTabs"
      form-focus
      grid-hide-paging
    >
      <template #grid_item_buttons_1="{ props, item }">
        <action-attachment
          :ref="
            (el) => {
              actAttchs.push(el);
            }
          "
          v-if="hasAttch"
          :kind="attchKind"
          :ref-id="attchRefId"
          :tags="[`${attchTagPrefix}_${attchRefId}_${item.LineNo}`]"
          :read-only="readOnly"
          @preOpen="emit('preOpenAttch', readOnly)"
          @close="emit('closeAttch', readOnly)"
        />
        <slot name="grid_item_buttons_1" :props="props" :item="item"> </slot>
      </template>
      <template #grid_Dimension="{ item, header }">
        <slot name="grid_Dimension" :item="item" :header="header">
          <DimensionText :dimension="item.Dimension" />
        </slot>
      </template>
      <template #grid_TagObjectID1="{ item, header }">
        <slot name="grid_TagObjectID1" :item="item" :header="header">
          <AccountSelector v-model="item.TagObjectID1" :read-only="readOnly" />
        </slot>
      </template>
      <template #grid_TagObjectID2="{ item, header }">
        <slot name="grid_TagObjectID2" :item="item" :header="header">
          <AccountSelector v-model="item.TagObjectID2" />
        </slot>
      </template>
      <template #grid_OffsetAccount="{ item, header }">
        <slot name="grid_OffsetAccount" :item="item" :header="header">
          <AccountSelector v-model="item.OffsetAccount" :read-only="readOnly" />
        </slot>
      </template>
      <template #grid_Account="{ item, header }">
        <slot name="grid_Account" :item="item" :header="header">
          <AccountSelector v-model="item.Account" :read-only="readOnly" />
        </slot>
      </template>
      <template #grid_PaymentType="{ item, header }">
        <slot name="grid_PaymentType" :item="item" :header="header"></slot>
      </template>
      <template #grid_ChequeGiroID="{ item, header }">
        <slot name="grid_ChequeGiroID" :item="item" :header="header"></slot>
      </template>
      <template #grid_Ignore="{ item, header }">
        <slot name="grid_Ignore" :item="item" :header="header"></slot>
      </template>
      <template #form_input_Dimension="{ item }">
        <slot name="form_input_Dimension" :item="item">
          <dimension-editor
            v-model="item.Dimension"
            :read-only="readOnly"
          ></dimension-editor>
        </slot>
      </template>
      <template #form_input_TagObjectID1="{ item }">
        <slot name="form_input_TagObjectID1" :item="item">
          <AccountSelector v-model="item.TagObjectID1" />
        </slot>
      </template>
      <template #form_input_TagObjectID2="{ item }">
        <slot name="form_input_TagObjectID2" :item="item">
          <AccountSelector v-model="item.TagObjectID2" />
        </slot>
      </template>
      <template #form_input_OffsetAccount="{ item }">
        <slot name="form_input_OffsetAccount" :item="item">
          <AccountSelector v-model="item.OffsetAccount" />
        </slot>
      </template>
      <template #form_input_Account="{ item }">
        <slot name="form_input_Account" :item="item">
          <AccountSelector v-model="item.Account"></AccountSelector>
        </slot>
      </template>
      <template #form_input_PaymentType="{ item }">
        <slot name="grid_PaymentType" :item="item"></slot>
      </template>
      <template #form_input_ChequeGiroID="{ item }">
        <slot name="grid_ChequeGiroID" :item="item"></slot>
      </template>
      <template #form_input_Ignore="{ item }">
        <slot name="grid_Ignore" :item="item"></slot>
      </template>

      <template #form_tab_References="{ item, mode }">
        <div class="w-full min-h-[400px]">
          <div v-if="data.loadingReferences"></div>
          <references-line
            v-show="!data.loadingReferences"
            v-model="item.References"
            :ref-list="data.refList"
            :hide-manual-input="referencesHideManualInput"
          />
        </div>
      </template>
      <template #form_tab_Checklist="{ item, mode }">
        <div class="w-full min-h-[400px]">
          <div v-if="data.loadingReferences"></div>
          <checklist-line
            v-show="!data.loadingReferences"
            v-model="item.ChecklistTemp"
            :attch-kind="attchKind"
            :attch-ref-id="item.LineNo"
            :attch-tag-prefix="attchTagPrefix"
          />
        </div>
      </template>
    </data-list>
  </div>
</template>
<script setup>
import { reactive, ref, watch, onMounted, inject, computed } from "vue";
import {
  DataList,
  util,
  SModal,
  SForm,
  SButton,
  SInput,
  createFormConfig,
} from "suimjs";

import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionText from "@/components/common/DimensionText.vue";
import AccountSelector from "@/components/common/AccountSelector.vue";
import ActionAttachment from "@/components/common/ActionAttachment.vue";
import ReferencesLine from "@/components/common/ReferencesLine.vue";
import ChecklistLine from "@/components/common/ChecklistLine.vue";
import helper from "@/scripts/helper.js";

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  readOnly: { type: Boolean, default: false },
  gridConfigUrl: {
    type: String,
    default: "/fico/customerjournal/line/gridconfig",
  },
  formConfigUrl: {
    type: String,
    default: "/fico/customerjournal/line/formconfig",
  },
  gridHideDetail: { type: Boolean, default: false },
  attchKind: {
    type: String,
    default: "",
  },
  attchRefId: {
    type: String,
    default: "",
  },
  attchTagPrefix: {
    type: String,
    default: "",
  },
  showReferences: {
    type: Boolean,
    default: false,
  },

  showChecklist: {
    type: Boolean,
    default: false,
  },
  referenceTemplate: { type: String, default: "" },
  checklistId: { type: String, default: "" },
  referencesHideManualInput: { type: Boolean, default: false },
});

const emit = defineEmits({
  "update:modelValue": null,
  newRecord: null,
  gridRowFieldChanged: null,
  calc: null,
  alterGridConfig: null,
  preOpenAttch: null,
  closeAttch: null,
  postDelete: null,
});

const axios = inject("axios");
const listControl = ref(null);
const actAttchs = ref([]);

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  records:
    props.modelValue == null || props.modelValue == undefined
      ? []
      : props.modelValue,
  formTabs: buildFormTabs(),
  refList: {
    Items: [],
  },
  checklist: [],
  loadingReferences: false,
  loadingChecklist: false,
  originRecord: {},
});

const hasAttch = computed({
  get() {
    return props.attchKind != "" && props.attchRefId != "";
  },
});
function buildFormTabs() {
  let r = ["General"];
  if (props.showReferences) r.push("References");
  if (props.showChecklist) r.push("Checklist");
  return r;
}

function openForm(r) {
  emit("preOpenAttch", props.readOnly);
  data.originRecord = helper.cloneObject(r);
  util.nextTickN(2, () => {
    if (props.readOnly === true) {
      data.formMode = "view";
      listControl.value.setFormMode("view");
    }
  });
}
function getMaxLineNo() {
  const obj = data.records.reduce(
    (acc, c) => {
      return acc.LineNo > c.LineNo ? acc : c;
    },
    { LineNo: 0 }
  );
  return obj.LineNo ?? 0;
}
function onGridNewRecord(r) {
  r.ID = "";
  r.Amount = 0;
  r.LineNo = getMaxLineNo() + 1;
  r.CurrencyID = "IDR";
  r.ChequeGiroID = "";
  r.Critical = false;
  r.References = mappingRefValues();
  r.ChecklistTemp = data.checklist;
  emit("newRecord", r);
  data.records.push(r);
  updateItems();
}

function onGridRowDelete(_, index) {
  const rowDelete = () => {
    const deletedRecord = data.records[index];
    const newRecords = data.records.filter((_, idx) => {
      return idx != index;
    });
    data.records = newRecords;
    util.nextTickN(2, () => {
      calc();
      updateItems();
      emit("postDelete", deletedRecord);
    });
  };
  if (hasAttch.value && actAttchs.value[index] !== undefined) {
    actAttchs.value[index].deleteAttch(rowDelete);
  } else {
    rowDelete();
  }
}

function calc() {
  const Total = data.records.reduce(
    (total, e) => {
      total.PriceEach += e.PriceEach;
      total.Qty += e.Qty;
      if (e.DiscountType == "fixed") {
        total.Discount += e.Discount;
      } else {
        const totalAmount = e.PriceEach * e.Qty;
        const discountAmount = totalAmount * (parseFloat(e.Discount) / 100);
        total.AmountPercent += parseInt(discountAmount);
      }
      total.Subtotal += e.PriceEach * e.Qty;
      total.Amount += e.Amount;
      return total;
    },
    {
      PriceEach: 0,
      Qty: 0,
      Discount: 0,
      AmountPercent: 0,
      Subtotal: 0,
      Amount: 0,
    }
  );
  emit("calc", Total);
}
function onAlterGridConfig(config) {
  emit("alterGridConfig", config);
  setTimeout(() => {
    updateItems();
    calc();
  }, 500);
}
function onGridRowFieldChanged(name, v1, v2, old, record) {
  if (name == "Qty") {
    record.Amount = v1 * record.PriceEach;
  }
  if (name == "PriceEach") {
    record.Amount = v1 * record.Qty;
  }
  if (name == "Discount") {
    if (record.DiscountType == "fixed") {
      const totalAmount = record.PriceEach * record.Qty;
      record.Amount = totalAmount - v1;
    } else {
      const totalAmount = record.PriceEach * record.Qty;
      const discountAmount = totalAmount * (parseFloat(v1) / 100);
      record.Amount = totalAmount - discountAmount;
    }
  }
  gridFieldChange(name, record);
  emit("gridRowFieldChanged", name, v1, v2, old, record, () => {
    gridFieldChange(name, record);
  });
}
function gridFieldChange(name, record) {
  listControl.value.setGridRecord(
    record,
    listControl.value.getGridCurrentIndex()
  );

  updateItems();

  if (
    ["Qty", "PriceEach", "DiscountType", "Discount", "Amount"].includes(name)
  ) {
    util.nextTickN(2, () => {
      calc();
    });
  }
}
function updateItems() {
  listControl.value.setGridRecords(data.records);
}
function refresh() {
  data.records = props.modelValue;
  util.nextTickN(2, () => {
    calc();
    updateItems();
  });
}

function onFormPostSave(params) {
  util.nextTickN(2, () => {
    calc();
  });
}
function getReferenceTemplate() {
  if (!props.showReferences) return;
  if (props.referenceTemplate == "") {
    data.refList = {
      Items: [],
    };
    return;
  }
  data.loadingReferences = true;
  const url = "/tenant/referencetemplate/get";

  axios
    .post(url, [props.referenceTemplate])
    .then(
      (r) => {
        data.records.forEach((e, i) => {
          if (
            data.records[i].References.length == 0 ||
            data.records[i].References == null
          )
            data.records[i].References = mappingRefValues(
              r.data.Items,
              e.References ?? []
            );
        });
        data.refList = r.data;
      },
      (e) => {
        data.refList = {
          Items: [],
        };
      }
    )
    .finally(() => {
      data.loadingReferences = false;
    });
}
function mappingRefValues(sources = data.refList.Items, references = []) {
  const defaultValue = {
    date: new Date(),
    number: 0,
  };
  const mv = references.filter((e) => e.isDefaultTemplate != true);

  const res = sources.map((el) => {
    const ky = el.Label;
    const f = mv.filter((o) => o.Key == ky);
    const v = f.length == 0 ? defaultValue[el.ReferenceType] ?? "" : f[0].Value;
    return {
      Key: ky,
      Value: v,
      isDefaultTemplate: f.length == 0,
    };
  });
  return [...res, ...mv];
}
function changeChecklistID() {
  if (!props.showChecklist) return;
  if (props.checklistId == "") {
    return;
  }
  data.loadingChecklist = true;
  axios
    .post("/tenant/checklisttemplate/get", [props.checklistId])
    .then(
      (r) => {
        data.checklist = r.data.Checklists;
        data.records.forEach((e, i) => {
          if (
            data.records[i].ChecklistTemp?.length == 0 ||
            data.records[i].ChecklistTemp == null
          ) {
            data.records[i].ChecklistTemp = r.data.Checklists;
          } else {
            let newRecords = [];
            e.ChecklistTemp?.forEach((el) => {
              const existings = data.records[i].ChecklistTemp.filter(
                (rcd) => rcd.Key == el.Key
              );
              if (existings.length == 0) {
                el.suimRecordChange = false;
                newRecords.push(el);
              }
            });
            e.ChecklistTemp = [...e.ChecklistTemp, ...newRecords];
          }
        });
      },
      (e) => {
        data.checklist = [];
      }
    )
    .finally((e) => {
      data.loadingChecklist = false;
    });
}
defineExpose({
  calc,
  refresh,
});

watch(
  () => data.records,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);

watch(
  () => props.referenceTemplate,
  (nv) => {
    getReferenceTemplate();
  },
  { deep: true }
);

watch(
  () => props.checklistId,
  (nv) => {
    changeChecklistID();
  },
  { deep: true }
);
onMounted(() => {
  getReferenceTemplate();
});
</script>
