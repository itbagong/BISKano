<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      title="Cheque & Giro Book"
      grid-config="/fico/cgbook/gridconfig"
      form-config="/fico/cgbook/formconfig"
      grid-read="/fico/cgbook/gets"
      form-read="/fico/cgbook/get"
      grid-mode="grid"
      grid-delete="/fico/cgbook/delete"
      form-keep-label
      form-insert="/fico/cg/create-update-book"
      form-update="/fico/cg/create-update-book"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      form-hide-submit
      :grid-fields="['CashBookID', 'Kind']"
      :form-fields="['Dimension']"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <template #grid_CashBookID="{ item }">
        <s-input
          v-model="item.CashBookID"
          hide-label
          use-list
          lookup-url="/tenant/cashbank/find"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          :read-only="true"
        />
      </template>
      <template #grid_Kind="{ item }">
        <s-input
          v-model="item.Kind"
          hide-label
          use-list
          lookup-url="/tenant/masterdata/find?MasterDataTypeID=CGT"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          :read-only="true"
        />
      </template>
      <template #form_buttons_1="{ item }">
        <s-button
          icon="cog-transfer"
          class="btn_primary"
          label="Generate"
          @click="onGenerate(item)"
        />
      </template>

      <template #grid_item_buttons="{ item }">
        <mdicon
          v-if="profile.canDelete"
          name="delete"
          width="16"
          alt="delete"
          class="cursor-pointer hover:text-primary"
          @click="onDelete(item)"
        />
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor-vertical v-model="item.Dimension"  :default-list="profile.Dimension" />
      </template>
    </data-list>

    <s-modal :display="false" ref="deleteModal" @submit="confirmDelete">
      You will delete data ! Are you sure ?<br />
      Please be noted, this can not be undone !
    </s-modal>
  </div>
</template>
<script setup>
import { reactive, ref, inject } from "vue";
import { DataList, util, SInput, SButton, SModal } from "suimjs";
import { layoutStore } from "@/stores/layout.js";
import { authStore } from "@/stores/auth.js";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";

layoutStore().name = "tenant";

const FEATUREID = 'Administrator'
const profile = authStore().getRBAC(FEATUREID)

const listControl = ref(null);
const deleteModal = ref(null);
const axios = inject("axios");

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  confirmDelete: false,
  deleteFn: undefined,
});

function onGenerate(record) {
  listControl.value.submitForm(
    record,
    () => {},
    () => {}
  );
}

function onDelete(record) {
  data.deleteFn = () => {
    const url = "/fico/cg/validate-delete-book";
    axios.post(url, record).then(
      (r) => {
        deleteData(record);
      },
      (e) => {
        util.showError(e);
      }
    );
  };
  deleteModal.value.show();
}

function confirmDelete() {
  deleteModal.value.hide();
  data.deleteFn();
}

function deleteData(record) {
  const url = "/fico/cgbook/delete";
  axios.post(url, record).then(
    (r) => {
      listControl.value.refreshGrid();
      deleteDetail(record);
    },
    (e) => {}
  );
}

function deleteDetail(record) {
  const url = "/fico/cg/delete-cheque-giro";
  axios.post(url, record).then(
    (r) => {},
    (e) => {}
  );
}
</script>
