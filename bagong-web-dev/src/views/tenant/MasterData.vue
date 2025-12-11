<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      :title="`Master Data ${data.title}`"
      grid-config="/tenant/masterdata/gridconfig"
      form-config="/tenant/masterdata/formconfig"
      :grid-read="data.gridread"
      form-read="/tenant/masterdata/get"
      grid-mode="grid"
      grid-delete="/tenant/masterdata/delete"
      form-keep-label
      form-insert="/tenant/masterdata/save"
      form-update="/tenant/masterdata/save"
      :grid-fields="['Enable']"
      :form-fields="['ParentID']"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="openForm"
      @preSave="onPreSave"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <template #form_input_ParentID="{ item }">
        <s-input
          :hide-label="false"
          label="Parent ID"
          v-model="item.ParentID"
          class="w-full"
          :useList="data.parentId != ''"
          :disabled="data.parentId == ''"
          :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=${data.parentId}`"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
        ></s-input>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, watch, onMounted, inject } from "vue";
import { authStore } from "@/stores/auth.js";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput } from "suimjs";
import { useRoute } from "vue-router";

layoutStore().name = "tenant";

const FEATUREID = "MasterData";
const profile = authStore().getRBAC(FEATUREID);

const listControl = ref(null);
const route = useRoute();
const axios = inject("axios");
const data = reactive({
  appMode: "grid",
  formMode: "edit",
  title: route.query.title || route.query.objname,
  masterDataTypeID: route.query.MasterDataTypeID,
  gridread: `/tenant/masterdata/gets?MasterDataTypeID=${route.query.MasterDataTypeID}`,
  parentId: "",
});

function newRecord(record) {
  record._id = "";
  record.Name = "";
  record.ParentID = "";
  record.IsActive = true;

  openForm(record);
}

function openForm() {
  util.nextTickN(2, () => {
    // listControl.value.setFormFieldAttr("_id", "rules", [
    //   (v) => {
    //     let vLen = 0;
    //     let consistsInvalidChar = false;
    //     v.split("").forEach((ch) => {
    //       vLen++;
    //       const validCar =
    //         "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz-_".indexOf(
    //           ch
    //         ) >= 0;
    //       if (!validCar) consistsInvalidChar = true;
    //       //console.log(ch,vLen,validCar)
    //     });
    //     if (vLen < 2 || consistsInvalidChar)
    //       return "minimal length is 2 and alphabet only";
    //     return "";
    //   },
    // ]);
  });
}

function getMasterType() {
  axios
    .post("/tenant/masterdatatype/get", [route.query.MasterDataTypeID])
    .then((r) => {
      const msData = r.data;
      data.parentId = msData.ParentID;
    });
}
function onPreSave(record) {
  record.MasterDataTypeID = route.query.MasterDataTypeID;
}
watch(
  () => route.query.MasterDataTypeID,
  (nv) => {
    data.gridread = `/tenant/masterdata/gets?MasterDataTypeID=${nv}`;
    util.nextTickN(2, () => {
      listControl.value.refreshList();
      listControl.value.refreshForm();
      getMasterType();
    });
  }
);

watch(
  () => route.query.title,
  (nv) => {
    data.title = nv;
    listControl.value.setControlMode("grid");
  }
);

onMounted(async () => {
  getMasterType();
});
</script>
