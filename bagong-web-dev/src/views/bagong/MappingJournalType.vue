<template> 
    <data-list
      class="card datalist-map-journal-type"
      ref="listControl"
      title="Mapping Site Journal Type"
      no-gap
      grid-editor
      gridHideDetail
      grid-hide-search
      grid-hide-sort
      grid-hide-footer
      gridHideNew
      grid-no-confirm-delete
      init-app-mode="grid"
      grid-mode="grid"
      form-keep-label
      grid-config="/tenant/siteentryjournaltype/gridconfig"
      form-config="/tenant/siteentryjournaltype/formconfig"
      :grid-fields="['JournalTypeID','SiteID','Type']"
      :form-default-mode="data.formMode"
      @grid-row-add="newRecord"
      @alterGridConfig="alterGridConfig"
      @grid-row-save="onGridRowSave"
      @grid-row-delete="onGridRowDelete"
      @gridRowFieldChanged="onGridRowFieldChanged"
      @gridRowUpdate="onGridRowUpdate"
      @gridRefreshed="gridRefreshed" 
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <template #grid_header_buttons_2="{ item }">
        <s-button
          v-if="profile.canCreate"
          class="btn_primary ml-2"
          icon="plus"
          label="Add"
          @click="newRecord"
        ></s-button>
      </template>
      <template #grid_SiteID="{ item }">
        <template v-if="!item.IsEdited">{{item.SiteID}}</template>
      </template>
      <template #grid_Type="{ item }">
        <template v-if="!item.IsEdited">{{item.Type}}</template>
      </template>
      <template #grid_JournalTypeID="{ item }">
        <template v-if="!item.IsEdited">{{item.JournalTypeID}}</template>
        <s-input
          v-else
          :key="item.Type"
          v-model="item.JournalTypeID"
          label="Journal Type"
          use-list
          :lookup-url="
            item.Type === 'Revenue'
              ? '/fico/customerjournaltype/find'
              : item.Type === 'Payroll'
              ? '/fico/ledgerjournaltype/find'
              : '/fico/vendorjournaltype/find'
          "
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-search="['Name']"
          @change="()=>{ item.suimRecordChange = true }"
        />
      </template>
      <template #grid_item_buttons_1="{item, config}">
          <a
            v-if="item.IsEdited == false && profile.canUpdate"
            href="#" 
            @click="item.IsEdited = true"
            class="edit_action"
          >
            <mdicon
              name="pencil"
              width="16"
              alt="edit"
              class="cursor-pointer hover:text-primary"
            />
          </a>
      </template>
      <template #grid_paging>&nbsp;</template>
    </data-list> 
</template>
<script setup>
import { reactive, ref, inject } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SButton, SInput } from "suimjs";
import { authStore } from '@/stores/auth';
const axios = inject("axios");

layoutStore().name = "tenant";

const FEATUREID = 'MappingJournalType'
const profile = authStore().getRBAC(FEATUREID)

const listControl = ref(null);

const data = reactive({
  keyComponenetJournalType: 0,
  appMode: "grid",
  formMode: profile.canUpdate ? 'edit' : 'view',
  records: [],
});

function alterGridConfig(cfg) {
  cfg.fields = cfg.fields.filter(
    (el) => ["IsActive", "_id"].indexOf(el.field) == -1
  );
}

function newRecord() {
  const record = {};
  record.Site = "";
  record.Type = "";
  record.JournalTypeID = "";
  record.suimRecordChange = true;
  record.IsEdited = true

  let data = listControl.value.getGridRecords();
  data.records = data;
  data.records.push(record);
  listControl.value.setGridRecords(data.records);
  util.nextTickN(2, () => {
    const el = document.querySelector(".datalist-map-journal-type .suim_area_table")
    el.scrollTop = el.scrollHeight
  })
}

function onGridRowSave(record, index) {
  let payload = record;
  axios.post("/tenant/siteentryjournaltype/save", payload).then(
    async (r) => {
      record.suimRecordChange = false;
      util.showInfo("Data has been save");
    },
    (e) => {
      record.suimRecordChange = true;
      util.showError(e);
    }
  );
}

function onGridRowDelete(records, index) {
  let payload = records;
  axios.post("/tenant/siteentryjournaltype/delete", payload.items[index]).then(
    async (r) => {
      const newRecords = data.records.filter((dt, idx) => {
        return idx != index;
      });
      data.records = newRecords;
      listControl.value.setGridRecords(data.records);
      util.showInfo("Data has been delete");
    },
    (e) => {
      util.showError(e);
    }
  );
}
function onGridRowFieldChanged(name, v1, v2, old, record) {
  
  if (name == "Type") {
    record.JournalTypeID = "";
  }
}

function gridRefreshed() {
  listControl.value.setGridLoading(true)
  axios.post("/tenant/siteentryjournaltype/gets", {}).then(
    (r) => {
      listControl.value?.setGridLoading(false)
      data.records = r.data.data.map(e=>{
        e.IsEdited = false
        return e
      });
      listControl.value.setGridRecords(data.records);
    },
    (e) => {
      listControl.value?.setGridLoading(false)   
      util.showError(e);
    }
  );
}
</script>
<style>
table td:last-child {
    text-align: center;
}
.datalist-map-journal-type .suim_area_table{
  position: relative;
  max-height:calc(100vh - 250px);
  overflow: auto;
}
.datalist-map-journal-type .suim_table > thead{
  
  position: sticky;
  top: 0;
}
</style>