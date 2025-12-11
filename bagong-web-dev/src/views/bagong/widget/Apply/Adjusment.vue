<template>
  <div class="border-[1px] p-2 items-center">
    <div class="flex justify-between bg-slate-100 p-1">
      <div class="font-semibold">Adjustment</div>
      <button
        v-if="!disabled"
        @click="onAction"
        class="w-[60px] h-[24px] flex justify-center items-center bg-primary text-white"
      >
        <template v-if="data.disableAction === true">
          <mdicon name="pencil" size="12" class="mr-1" /> Edit
        </template>
        <template v-else>
          <mdicon name="content-save" size="12" class="mr-1" /> Save
        </template>
      </button>
    </div>
    <div class="grid grid-cols-4 gap-4 mt-2">
      <div
        v-for="(adj, idx) in data.items"
        :key="idx"
        class="flex gap-1 items-center"
      >
        <div>
          <s-input
            :read-only="data.disableAction"
            v-model="adj.Amount"
            :label="adj.Text == '' ? 'Adj ' + parseInt(idx + 1) : adj.Text"
            kind="number"
          />
          <span
            v-if="!data.disableAction"
            class="text-[0.7rem] text-primary underline-offset-2"
            @click="openForm(idx)"
            >Detail...</span
          >
        </div>
        <button
          v-if="!data.disableAction"
          @click="onAction"
          class="-mt-1 w-[24px] h-[24px] flex justify-center items-center text-primary"
        >
          <mdicon name="close-thick" size="20" />
        </button>
      </div>
      <button
        v-if="!data.disableAction"
        @click="openForm()"
        class="w-[45px] h-[24px] mt-4 flex justify-center items-center bg-primary text-white"
      >
        <mdicon name="plus" size="12" /> Adj
      </button>
    </div>
    <s-modal
      :display="data.modalFrm"
      hideButtons
      title="Adjustment"
      @beforeHide="data.modalFrm = false"
    >
      <div class="min-w-[500px]">
        <s-form
          v-model="data.frmRecord"
          :config="data.frmCfg"
          keep-label
          hide-cancel
          buttons-on-bottom
          :buttons-on-top="false"
          ref="frmCtl"
          @submit-form="onSave"
        >
          <template #input_Account="{ item, config }">
            <AccountSelector v-model="item.Account" />
          </template>
        </s-form>
      </div>
    </s-modal>
  </div>
</template>
<script setup>
import { reactive, computed, inject, ref, onMounted } from "vue";
import { SInput, SForm, loadFormConfig, util, SModal } from "suimjs";

import AccountSelector from "@/components/common/AccountSelector.vue";
const props = defineProps({
  readOnly: { type: Boolean, default: false },
  selectedSourceId: { type: String, default: "" },
  adjusments: { type: Object, default: () => [] },
  sourceRecId: { type: String, default: "" },
});
const emit = defineEmits({
  submit: null,
});
const axios = inject("axios");
const frmCtl = ref(null);

const data = reactive({
  disableAction: true,
  items:
    props.adjusments === undefined
      ? []
      : JSON.parse(JSON.stringify(props.adjusments)),
  modalFrm: false,
  frmCfg: {},
  frmRecord: {},
  idxItem: null,
});
const disabled = computed({
  get() {
    if (props.readOnly === true) return true;
    if (props.selectedSourceId === "") return true;
    else if (
      props.sourceRecId != "" &&
      props.sourceRecId !== props.selectedSourceId
    )
      return true;
    else return false;
  },
});
function onAction() {
  data.disableAction = !data.disableAction;
  if (data.disableAction === true)
    emit("submit", data.items, function () {
      data.items =
        props.adjusments === undefined
          ? []
          : JSON.parse(JSON.stringify(props.adjusments));
    });
}
function openForm(idx) {
  loadFormConfig(axios, "/fico/applyadjustment/line/formconfig").then(
    (r) => {
      data.frmCfg = r;
      util.nextTickN(2, () => {
        data.modalFrm = true;
        data.idxItem = idx;
        data.frmRecord =
          data.idxItem === undefined || data.idxItem === null
            ? {
                Text: "",
                TrxType: "ApplyAdjustment",
                Amount: 0,
              }
            : data.items[data.idxItem];
      });
    },
    (e) => util.showError(e)
  );
}
function onSave() {
  if (data.idxItem == undefined || data.idxItem === null)
    data.items.push(data.frmRecord);
  else data.items[daa.idxItem] = data.frmRecord;

  data.modalFrm = false;
}
function onRemove(idx) {}
onMounted(() => {});
</script>