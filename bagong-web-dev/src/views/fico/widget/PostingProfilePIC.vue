<template>
  <div>
    <data-list
      ref="dataCtl"
      title="Posting Profile"
      no-gap
      :grid-hide-delete="readOnly"
      :grid-hide-control="readOnly"
      grid-hide-search
      grid-hide-sort
      grid-hide-select
      grid-config="/fico/postingprofile/pic/gridconfig"
      form-config="/fico/postingprofile/pic/formconfig"
      :grid-read="`/fico/postingprofile/pic/get-by-posting-profile-id?PostingProfileID=${profile._id}`"
      form-read="/fico/postingprofile/pic/get"
      grid-mode="grid"
      grid-delete="/fico/postingprofile/pic/delete"
      grid-sort-field="Priority"
      grid-sort-direction="asc"
      form-keep-label
      form-insert="/fico/postingprofile/pic/insert"
      form-update="/fico/postingprofile/pic/update"
      :form-fields="['Dimension', 'Approvers', 'Account']"
      :grid-fields="['Dimension']"
      form-focus
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="openForm"
      :grid-hide-edit="readOnly"
      @preSave="preSave"
    >
      <template #form_input_Dimension="{ item, mode }">
        <dimension-editor
          v-model="item.Dimension"
          :dimNames="['PC', 'CC', 'Site', 'Asset']"
          :read-only="mode == 'view'"
        ></dimension-editor>
      </template>
      <template #form_footer_1="{ item, mode }">
        <div class="flex flex-col gap-6">
          <PostingProfileApprovers
            v-model="item.Approvers"
            title="Approvers"
            :read-only="mode == 'view'"
          />
          <PostingProfileApprovers
            v-model="item.Submitters"
            title="Submitters"
            :read-only="mode == 'view'"
          />
          <PostingProfileApprovers
            v-model="item.Postingers"
            title="Postingers"
            :read-only="mode == 'view'"
          />
        </div>
      </template>
      <template #grid_Dimension="{ item, mode }">
        <DimensionText
          :dimension="item.Dimension"
          :read-only="mode == 'view'"
        />
      </template>
      <template #form_input_Account="{ item, mode }">
        <AccountSelector
          v-model="item.Account"
          multiple-account-id
          :read-only="mode == 'view'"
        />
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref } from "vue";
import { DataList } from "suimjs";
import DimensionEditor from "@/components/common/DimensionEditor.vue";
import DimensionText from "@/components/common/DimensionText.vue";
import PostingProfileApprovers from "./PostingProfileApprovers.vue";
import AccountSelector from "@/components/common/AccountSelector.vue";

const props = defineProps({
  profile: { type: Object, default: () => {} },
  readOnly: { type: Boolean, default: false },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const dataCtl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: "edit",
});

function newRecord(record) {
  record._id = "";
  record.Priority = 1;
  record.PostingProfileID = props.profile._id;
}

function preSave(record) {
  console.log(record);
}
</script>
