<template> 
  <s-card
    title="Non Asset Entry"
    class="w-full bg-white suim_datalist"
    hide-footer
  >
    <s-form
      ref="frmCtl"
      v-model="data.record"
      :config="data.formCfg"
      :mode="data.formMode"
      keep-label
      :buttons-on-bottom="false"
      buttons-on-top
      @submit-form="onSaveForm"
      @cancelForm="emit('close')"
      :hideSubmit="readOnly"
    > 
      <template #buttons_1="{ item, inSubmission,loading }">
        <s-button
          :disabled="inSubmission || loading"    
          v-if="data.haveDraft"
          @click="onSubmit(item._id)"
          class="btn_primary"
          label="Submit"
        ></s-button>
      </template>
      <template #form_header="{ item }">
        <div class="w-full border flex mb-4 items-center">
          <div class="basis-1/2 p-2">
            <div class="text-xs">
              {{ moment(item.Created).format("DD MMM YYYY") }}
            </div>
          </div>
          <div class="basis-1/2 grid grid-cols-3 p-2 gap-2 content-center">
            <div class="flex border-r">
              <div class="font-bold">Income :</div>
              <div class="pl-2">
                {{ util.formatMoney(item.Income, {}) }}
              </div>
            </div>
            <div class="flex">
              <div class="font-bold">Expense :</div>
              <div class="pl-2">
                {{ util.formatMoney(item.Expense, {}) }}
              </div>
            </div>
            <div class="flex border-r">
              <div class="font-bold">Revenue :</div>
              <div class="pl-2">
                {{ util.formatMoney(item.Revenue, {}) }}
              </div>
            </div>
          </div>
        </div>
      </template>
      <template #input_ExpenseDetail="{ item }">
        <div class="mt-5">
          <expense
            v-model="item.ExpenseDetail"
            @calc="calcLineTotal"  
            :grid-config-url="
              readOnly
                ? '/bagong/siteexpense-read/grid/gridconfig'
                : '/bagong/siteexpense/grid/gridconfig'
            "
            :group-id-value="props.groupIdValue"
            :attch-kind="data.attchKind" 
            :attch-ref-id="item._id"
            :attch-tag-prefix="data.attchKind"
            @preOpenAttch="preOpenAttch"
            @reOpen="reOpen" 
          />
        </div>
      </template>
    </s-form>
  </s-card>
</template>

<script setup>
import { reactive, ref, onMounted, inject, watch, computed } from "vue";
import {
  DataList,
  SInput,
  util,
  SButton,
  SCard,
  SForm,
  loadFormConfig,
} from "suimjs";
import { authStore } from "@/stores/auth";
import helper from "@/scripts/helper.js";
import HeaderInfo from "./HeaderInfo.vue";

import Expense from "./Expense.vue";

import moment from "moment";
const props = defineProps({
  nonAssetId: { type: String, default: "" },
  siteID: { type: String, default: "" },
  groupIdValue: { type: Array, default: () => [] },
});

const emit = defineEmits({
  "update:modelValue": null,
  disableAction: null,
  close: null,
});

const auth = authStore();

const axios = inject("axios");
const frmCtl = ref(null);

const data = reactive({
  appMode: "grid",
  record: {
    ExpenseDetail: [],
  },
  formCfg: {},
  loading: false,
  readOnly: false,
  status: "DRAFT",
  disabledFormButton: false,  
  attchKind: "SE_NON_ASSET", 
  haveDraft: false,
  haveSubmitted: false,
  haveNewLine: false,
});
watch(() => data.record.ExpenseDetail, (nv) => {
  let found = false 
  nv.forEach(e=>{ 
    if(found) return;
    if(e.JournalID == '' || e.JournalID == undefined) found = true
  })
  data.haveDraft = found 
}, { deep: true })

const readOnly = computed({
  get() {
    return data.haveSubmitted;
  },
});
function calcLineTotal(total = { Value: 0, Amount: 0, TotalAmount: 0 }) {
  data.record.Expense = total.TotalAmount;
  data.record.Revenue = 0;
}
function preOpenAttch(){
  frmCtl.value.submit()
}
function fetchNonAsset() {
  data.status = "DRAFT";
  axios
    .post("/bagong/siteentry_nonasset/get", [props.nonAssetId])
    .then(
      (r) => {
        openForm(r.data);
      },
      (e) => {
        util.showError(e);
      }
    )
    .finally(() => {
      data.expKey++;
      loadFormConfig(axios, "/bagong/siteentry_nonasset/formconfig").then(
        (r) => {
          data.formCfg = r;
        },
        (e) => util.showError(e)
      );
    });
}
function checkLineStatus(lines){
  if(lines.length == 0) { 
    data.haveDraft = true
    data.haveSubmitted = false
  }else{
    const r = lines.filter(e=> e.JournalID == '')  
    data.haveDraft = r.length > 0
    const submitteds = lines.filter(e=> e.JournalID != '');
    data.haveSubmitted = submitteds.length > 0
  }
}
function openForm(r) {
  // console.log(r);
  data.record = r;
  checkLineStatus(data.record.ExpenseDetail)
}
function onSaveForm(record, cbSuccess, cbError) {
  save(cbSuccess,cbError)
}
function save(cbSuccess = () => {}, cbError = () => {}) {
  const param = data.record;
  axios.post("/bagong/siteentry_nonasset/save", param).then(
    (r) => {
      cbSuccess(r.data);
    },
    (e) => {
      cbError();
      util.showError(e);
    }
  );
}
function reOpen(){ 
  const cbSuccess = ()=>{
    checkLineStatus(data.record.ExpenseDetail)
    setFormLoading(false)
  }
  const cbError = ()=> {
    setFormLoading(false) 
  }
  setFormLoading(true)

  save(cbSuccess,cbError)
}
function setFormLoading(loading) { 
  frmCtl.value.setLoading(loading);
}
function onSubmit() {
  const validExpense = helper.validateSiteEntryExpense(data.record.ExpenseDetail)
  if (validExpense) { 
    setFormLoading(true);
    save(doSubmit);
  }
}
function doSubmit() {
  const url = "/bagong/postingprofile/post";
  const param = {
    JournalType: "SITEENTRY_NONASSET",
    JournalID: props.nonAssetId,
    Op: "Submit",
    Text: "",
  };
  axios.post(url, param).then(
    (r) => {
      setFormLoading(false);
      emit("close");
    },
    (e) => {
      util.showError(e);
      setFormLoading(false);
    }
  );
}

defineExpose({
  onSubmit,
});

onMounted(() => {
  fetchNonAsset();
});
</script>
