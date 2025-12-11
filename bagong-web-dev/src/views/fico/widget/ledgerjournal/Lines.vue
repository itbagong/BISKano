<template>
  <div class="flex flex-col gap-2">
    <data-list
      ref="listControl"
      title="Posting Profile"
      hide-title
      no-gap
      :grid-hide-detail="readOnly"
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
      form-default-mode="view"
      grid-mode="grid"
      form-keep-label
      new-record-type="grid"
      grid-config="/fico/ledgerjournal/line/gridconfig"
      form-config="/fico/ledgerjournal/line/formconfig"
      :form-fields="['References', 'Dimension']"
      :grid-fields="[
        'TagObjectID1',
        'TagObjectID2',
        'Account',
        'OffsetAccount',
        'Dimension',
      ]"
      :grid-must-editable="true"
      :grid-must-fields="['Account', 'OffsetAccount', 'Text']"
      @grid-row-add="newRecord"
      @form-field-change="onFormFieldChanged"
      @form-edit-data="openForm"
      @grid-row-delete="onGridRowDelete"
      @grid-row-field-changed="onGridRowFieldChanged"
      @grid-refreshed="gridRefreshed"
      @post-save="onFormPostSave"
      form-focus
    >
      <template #grid_Account="{ item }">
        <div class="flex flex-col gap-2">
          <AccountSelector
            v-model="item.Account"
            hide-label
            :read-only="
              readOnly &&
              status != 'DRAFT' &&
              !(ledgerEditor && status == 'READY')
            "
            @change="(account) => onChangeAccountSelector(account, item)"
          />
        </div>
      </template>
      <template #grid_OffsetAccount="{ item }">
        <div class="flex flex-col gap-2">
          <AccountSelector
            v-model="item.OffsetAccount"
            hide-label
            :read-only="
              readOnly &&
              status != 'DRAFT' &&
              !(ledgerEditor && status == 'READY')
            "
          />
        </div>
      </template>
      <template #form_input_References="props">
        <References
          ReferenceTemplate="OK"
          :readOnly="readOnly || status != 'DRAFT'"
          v-model="props.item.References"
        />
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor
          v-model="item.Dimension"
          :readOnly="readOnly || status != 'DRAFT'"
        ></dimension-editor>
      </template>
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
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, watch, computed, inject } from "vue";
import { DataList } from "suimjs";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import AccountSelector from "@/components/common/AccountSelector.vue";
import References from "@/components/common/References.vue";
import { authStore } from "@/stores/auth.js";
import { readonly } from "vue";
import ActionAttachment from "@/components/common/ActionAttachment.vue";

const ledgerEditor = authStore().getRBAC("EditLedgerBeforePost").canSpecial1;
const axios = inject("axios");

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  status: { type: String, default: "" },
  readOnly: { type: Boolean, default: false },
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
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
  preOpenAttch: null,
  closeAttch: null,
  postDelete: null,
});

const listControl = ref(null);
const actAttchs = ref([]);

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  records: props.modelValue.map((dt) => {
    dt.suimRecordChange = false;
    return dt;
  }),
});
const hasAttch = computed({
  get() {
    return props.attchKind != "" && props.attchRefId != "";
  },
});
function getMaxLineNo() {
  const obj = data.records.reduce(
    (acc, c) => {
      return acc.LineNo > c.LineNo ? acc : c;
    },
    { LineNo: 0 }
  );
  return obj.LineNo ?? 0;
}
function newRecord() {
  const record = {};
  record.suimRecordChange = false;
  record.LineNo = getMaxLineNo() + 1;
  record.Amount = 1;
  record.References = [];
  record.Debit = 0;
  record.Credit = 0;
  data.records.push(record);
  listControl.value.setGridRecords(data.records);
  updateItems();
}

function openForm(record) {
  updateJournalType(record.JournalTypeID);
}

function onGridRowDelete(record, index) {
  const rowDelete = () => {
    const newRecords = data.records.filter((dt, idx) => {
      return idx != index;
    });
    data.records = newRecords;
    listControl.value.setGridRecords(data.records);
    updateItems();
    emit("postDelete");
  };

  if (hasAttch.value && actAttchs.value[index] !== undefined) {
    actAttchs.value[index].deleteAttch(rowDelete);
  } else {
    rowDelete();
  }
}

function onFormPostSave(record, index) {
  record.suimRecordChange = false;
  if (listControl.value.getFormMode() == "new") {
    data.records.push(record);
  } else {
    data.records[index] = record;
  }
  listControl.value.setGridRecords(data.records);
  updateItems();
}

function onFormFieldChanged(name, v1, v2, old, record) {
  if (name == "JournalTypeID") updateJournalType(v1);
  updateItems();
}

function updateJournalType(id) {}
function onGridRowFieldChanged(name, v1, v2, old, record) {
  if (name == "Debit") {
    record.Amount = v1 - record.Credit;
  }
  if (name == "Credit") {
    record.Amount = record.Debit - v1;
  }

  listControl.value.setGridRecord(
    record,
    listControl.value.getGridCurrentIndex()
  );
  updateItems();
}

function onChangeAccountSelector(v, item) {
  if (v.AccountType == "ASSET") {
    axios.post("/tenant/asset/find", {
      Take: 1,
      Where: {
        Field: "_id",
        Op: "$eq",
        Value: v.AccountID,
      },
    }).then(
    (r) => {
      item.Dimension = r.data[0]?.Dimension
      updateItems();
    },
    (e) => {
      util.showError(e)
    })
  }
}
function updateItems() {
  const committedRecords = data.records.filter(
    (dt) => dt.suimRecordChange == false || dt.suimRecordChange == undefined
  );
  emit("update:modelValue", committedRecords);
  emit("recalc");
}

function gridRefreshed() {
  listControl.value.setGridRecords(data.records);
}

watch(
  () => props.modelValue,
  (nv) => {
    if (nv.length !== data.records.length) {
      data.records = nv;
      gridRefreshed();
    }
  }
);
</script>
