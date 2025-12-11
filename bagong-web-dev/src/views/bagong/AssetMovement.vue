<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      title="Asset Movement"
      grid-config="/bagong/asset-movement/gridconfig"
      form-config="/bagong/asset-movement/formconfig"
      grid-read="/bagong/asset-movement/gets"
      form-read="/bagong/asset-movement/get"
      grid-mode="grid"
      grid-delete="/bagong/asset-movement/delete"
      form-insert="/bagong/asset-movement/save"
      form-update="/bagong/asset-movement/save"
      :form-fields="['TrxDate']"
      :init-app-mode="data.appMode"
      :form-tabs-edit="['General', 'Line']"
      :form-tabs-new="['General', 'Line']"
      @formFieldChange="onFormFieldChange"
      @formEditData="editRecord"
      @formNewData="newData"
      form-keep-label
      grid-hide-new
      v-if="!data.isAddNew"
      stay-on-form-after-save
    >
      <template #form_buttons_1="{ item, inSubmission, loading }">
        <form-buttons-trx
          :disabled="loading"
          :status="item.Status"
          :journal-id="item._id"
          :posting-profile-id="item.PostingProfileID"
          :journal-type-id="'ASSETMOVEMENT'"
          :moduleid="'bagong'"
          @preSubmit="trxPreSubmit"
          @postSubmit="trxPostSubmit"
          @errorSubmit="trxErrorSubmit"
          :auto-post="!waitTrxSubmit"
        />
      </template>
      <template #form_input_TrxDate="{ item }">
        <label class="input_label"><div>Trasaction Date</div></label>
        {{ moment(item.TrxDate).format("DD MMM yyyy") }}
      </template>
      <template #grid_header_buttons_2="{ item, mode }">
        <s-button
          icon="plus"
          class="btn_primary new_btn"
          tooltip="add new"
          @click="data.isAddNew = true"
        />
      </template>
      <template #form_tab_Line="{ item, mode }">
        <data-list
          ref="gridLine"
          grid-config="/bagong/asset-movement/line/gridconfig"
          grid-mode="grid"
          init-app-mode="grid"
          grid-hide-new
          grid-hide-refresh
          grid-hide-sort
          grid-hide-search
          grid-hide-edit
          grid-hide-delete
          grid-hide-select
          no-gap
          grid-hide-control
          @alter-grid-config="onAltergridCfg"
          grid-editor
          grid-auto-commit-line
          :grid-fields="['CustomerFrom', 'ProjectFrom', 'SiteFrom']"
        >
          <template #grid_CustomerFrom="{ item }">
            <s-input
              v-model="item.CustomerFrom"
              useList
              lookup-key="_id"
              :lookup-labels="['Name']"
              :lookupSearchs="['_id', 'Name']"
              lookupUrl="/tenant/customer/find"
              read-only
            />
          </template>
          <template #grid_ProjectFrom="{ item }">
            <s-input
              v-model="item.ProjectFrom"
              useList
              lookup-key="_id"
              :lookup-labels="['ProjectName']"
              :lookupSearchs="['_id', 'ProjectName']"
              lookupUrl="/sdp/measuringproject/find"
              read-only
            />
          </template>
          <template #grid_SiteFrom="{ item }">
            <s-input
              v-model="item.SiteFrom"
              useList
              lookup-key="_id"
              :lookup-labels="['Name']"
              :lookupSearchs="['_id', 'Name']"
              lookupUrl="/bagong/sitesetup/find"
              read-only
            />
          </template>
        </data-list>
      </template>
    </data-list>

    <select-asset
      v-if="data.isAddNew"
      @cancel="data.isAddNew = false"
      @movedAsset="onMovedAsset"
    />
  </div>
</template>
<script setup>
import { reactive, ref, watch, inject, onMounted, computed } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput, SButton } from "suimjs";
import SelectAsset from "./widget/MovementAssetSelect.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";

import moment from "moment";

const axios = inject("axios");
const listControl = ref(null);
const gridLine = ref(null);

layoutStore().name = "tenant";

const data = reactive({
  appMode: "grid",
  record: {},
  readOnly: false,
  sites: {},
  isAddNew: false,
});

function onMovedAsset(params) {
  data.isAddNew = false;
  axios
    .post("/fico/assetjournaltype/find", {
      Take: 20,
      Sort: ["Name"],
      Select: ["Name", "_id", "PostingProfileID"],
    })
    .then(
      (res) => {
        let r = {
          TrxDate: new Date(),
          Total: params.length ?? 0,
          Lines: params,
          JournalTypeID: res.data.length > 0 ? res.data[0]._id : "",
          PostingProfileID:
            res.data.length > 0 ? res.data[0].PostingProfileID : "",
        };
        util.nextTickN(3, () => {
          listControl.value.setControlMode("form");
          setTimeout(() => {
            listControl.value.setFormRecord(r);
            listControl.value.setFormCurrentTab(1);
            gridLine.value.setGridRecords(params);
          }, 500);
        });
      },
      (e) => {
        util.showError(e);
      }
    )
    .finally(() => {});
}
function getJurnalType(id) {
  if (id === "" || id === null) {
    data.record.PostingProfileID = "";
    return;
  }
  listControl.value.setFormLoading(true);
  axios
    .post("/fico/assetjournaltype/get", [id])
    .then(
      (r) => {
        data.record.PostingProfileID = r.data.PostingProfileID;
      },
      (e) => {
        data.record.PostingProfileID = "";
        util.showError(e);
      }
    )
    .finally(() => {
      listControl.value.setFormLoading(false);
    });
}
function onFormFieldChange(name, v1, v2, old, record) {
  switch (name) {
    case "JournalTypeID":
      getJurnalType(v1, record);
      break;
  }
}
function onAltergridCfg(cfg) {
  let noField = [
    "PCFrom",
    "PCTo",
    "NoHullCustomer",
    "DateFrom",
    "DateTo",
    "IsChecked",
    "CustomerFromName",
  ];
  if (data.record.Status !== "DRAFT") {
    noField = ["PCFrom", "PCTo", "IsChecked", "CustomerFromName"];
  }

  let editable = ["SiteTo", "CustomerTo", "ProjectTo"];
  if (data.record.Status === "SUBMITTED") {
    editable = ["NoHullCustomer", "DateFrom", "DateTo"];
  }

  let f = cfg.fields.filter((x) => {
    if (!editable.includes(x.field)) x.input.readOnly = true;
    return !noField.includes(x.field);
  });

  cfg.fields = f;
}
function setLoadingForm(loading) {
  listControl.value.setFormLoading(loading);
}
function setFormRequired(required) {
  listControl.value.getFormAllField().forEach((e) => {
    listControl.value.setFormFieldAttr(e.field, "required", required);
  });
}
const waitTrxSubmit = computed({
  get() {
    return ["DRAFT", "READY"].includes(data.record.Status);
  },
});

function trxPreSubmit(status, action, doSubmit) {
  if (waitTrxSubmit.value) {
    trxSubmit(doSubmit);
  }
}

function trxSubmit(doSubmit) {
  setFormRequired(true);
  util.nextTickN(2, () => {
    const valid = listControl.value.formValidate();
    if (valid) {
      setLoadingForm(true);
      listControl.value.submitForm(
        data.record,
        () => {
          doSubmit();
        },
        () => {
          setLoadingForm(false);
        }
      );
    }
    setFormRequired(false);
  });
}

function trxPostSubmit(record) {
  setLoadingForm(false);
  setModeGrid();
}
function setModeGrid() {
  listControl.value.setControlMode("grid");
  listControl.value.refreshList();
}
function trxErrorSubmit(e) {
  setLoadingForm(false);
}
function editRecord(r) {
  data.record = r;
  setTimeout(() => {
    util.nextTickN(3, () => {
      gridLine.value.setGridRecords(r.Lines);
    });
  }, 500);
}

onMounted(() => {});
</script>
