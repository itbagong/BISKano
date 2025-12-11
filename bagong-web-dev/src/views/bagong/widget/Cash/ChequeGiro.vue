<template>
  <div class="">
    <s-modal
      :display="data.modalApplyCheque"
      hideButtons
      title="Apply Cheque & Giro"
      @beforeHide="data.modalApplyCheque = false"
    >
      <s-card class="rounded-md w-full" hide-title>
        <s-form
          ref="frmApplyCheque"
          v-model="data.frmApplyCheque"
          :config="data.cfgApplyCheque"
          keep-label
          only-icon-top
          hide-submit
          hide-cancel
        >
          <template #input_CashBookID="{ item }">
            <s-input
              v-model="item.CashBookID"
              label="Bank Account No"
              use-list
              lookup-url="/tenant/cashbank/find"
              lookup-key="_id"
              :lookup-labels="['Name']"
              :lookup-searchs="['_id', 'Name']"
              :read-only="true"
            />
          </template>
          <template #input__id="{ item }">
            <s-input
              label="Cheque/Giro no"
              v-model="item._id"
              use-list
              lookup-url="/fico/cheque/find?Status=Open"
              lookup-key="_id"
              :lookup-labels="['_id']"
              :lookup-searchs="['_id']"
            />
          </template>
          <template #input_Kind="{ item }">
            <s-input
              label="Payment Type"
              v-model="item.Kind"
              use-list
              lookup-url="/tenant/masterdata/find?MasterDataTypeID=PTY"
              lookup-key="_id"
              :lookup-labels="['Name']"
              :lookup-searchs="['_id', 'Name']"
              :read-only="true"
            />
          </template>
          <template #footer_1="{}">
            <div class="w-full flex justify-end">
              <s-button
                name="Reserve"
                alt="Reserve"
                class="btn_primary cursor-pointer mt-[-3px]"
                width="16"
                label="Reserve"
                @click="onReserve"
              />
            </div>
          </template>
        </s-form>
      </s-card>
    </s-modal>
  </div>
</template>
<script setup>
import { reactive, ref, watch, onMounted, inject } from "vue";
import {
  DataList,
  util,
  SModal,
  SInput,
  createFormConfig,
  SForm,
  SButton,
} from "suimjs";
const axios = inject("axios");

const props = defineProps({
  journal: { type: Object, default: {} },
  lines: { type: Array, default: [] },
});

const data = reactive({
  modalApplyCheque: false,
  frmApplyCheque: {},
  cfgApplyCheque: {},
});

function genCfgApplyGiro() {
  const cfg = createFormConfig("", true);
  cfg.addSection("Bank", false).addRowAuto(1, {
    field: "CashBookID",
    label: "Bank Account No",
    kind: "text",
    readOnly: true,
  });

  cfg.addSection("General", false).addRowAuto(
    2,
    {
      field: "Kind",
      label: "Payment Type",
      kind: "text",
      useList: true,
      lookupKey: "_id",
      lookupUrl: "/tenant/masterdata/find?MasterDataTypeID=PTY",
      lookupLabels: ["Name"],
      lookupSearchs: ["_id", "Name"],
      readOnly: true,
    },
    {
      field: "_id",
      label: "Cheque/Giro No.",
      required: true,
      kind: "text",
      useList: true,
      lookupKey: "_id",
      lookupUrl: "/fico/cheque/find?Status=Open",
      lookupLabels: ["_id"],
      lookupSearchs: ["_id"],
    },
    {
      field: "IssueDate",
      label: "Issue Date",
      required: true,
      kind: "date",
    },
    {
      field: "ClearDate",
      label: "Clear Date",
      required: true,
      kind: "date",
    },
    {
      field: "Amount",
      label: "Amount",
      kind: "text",
      readOnly: true,
    },
    {
      field: "BfcName",
      label: "Bfc Name",
      kind: "text",
    }
  );

  cfg.addSection("Memo", false).addRowAuto(1, {
    field: "Memo",
    label: "Memo",
    required: true,
    kind: "text",
    multiRow: "3",
  });
  data.cfgApplyCheque = cfg.generateConfig();
}

function onApplyCheque(record) {
  data.modalApplyCheque = true;
  data.frmApplyCheque = {};
  data.frmApplyCheque = record;
  data.frmApplyCheque.CashBookID = props.journal.CashBookID;
  data.frmApplyCheque.Kind = record.PaymentType;
  data.frmApplyCheque.CashJournalID = props.journal._id;
  data.frmApplyCheque.IssueDate = new Date();
  data.frmApplyCheque.ClearDate = new Date();
}

function onReserve() {
  let f = props.lines.filter((o) => o.ChequeGiroID == data.frmApplyCheque._id);
  if (f.length > 0) {
    util.showError("Cheque already used");
    return;
  }
  data.frmApplyCheque.ChequeGiroID = data.frmApplyCheque._id;
  data.modalApplyCheque = false;
}

onMounted(() => {
  genCfgApplyGiro();
});

defineExpose({
  onApplyCheque,
});
</script>
