<template>
  <div class="w-full">
    <data-list
      class="card"
      ref="listControl"
      title="Posting Profile"
      grid-config="/fico/postingprofile/gridconfig"
      form-config="/fico/postingprofile/formconfig"
      grid-read="/fico/postingprofile/gets"
      form-read="/fico/postingprofile/get"
      grid-mode="grid"
      grid-delete="/fico/postingprofile/delete"
      form-keep-label
      form-insert="/fico/postingprofile/insert"
      form-update="/fico/postingprofile/update"
      :grid-fields="['Enable']"
      :form-tabs-edit="['General', 'Approvers']"
      :form-tabs-view="['General', 'Approvers']"
      :form-fields="['DirectPosting', 'DoNotPost', 'Dimension']"
      :init-app-mode="data.appMode"
      :form-default-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="openForm"
      @form-field-change="onFormFieldChange"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
    >
      <template #form_tab_Approvers="{ item, mode }">
        <PostingProfilePIC :profile="item" :read-only="mode == 'view'" />
      </template>
      <template #form_input_DirectPosting="{ item }">
        <div class="suim_input checkboxOffset">
          <div class="flex gap-2 items-center mt-1">
            <input
              type="checkbox"
              v-model="item.DirectPosting"
              :disabled="data.disableDirectPosting"
              @change="
                (e) => {
                  data.disableDoNotPost = e.target.checked;
                }
              "
            />
            <div>Direct posting</div>
          </div>
        </div>
      </template>
      <template #form_input_DoNotPost="{ item }">
        <div class="suim_input checkboxOffset">
          <div class="flex gap-2 items-center mt-1">
            <input
              type="checkbox"
              v-model="item.DoNotPost"
              :disabled="data.disableDoNotPost"
              @change="
                (e) => {
                  data.disableDirectPosting = e.target.checked;
                }
              "
            />
            <div>Do not post</div>
          </div>
        </div>
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor
          v-model="item.Dimension"
          :default-list="profile.Dimension"
        ></dimension-editor>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref } from "vue";
import { authStore } from "@/stores/auth.js";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util } from "suimjs";
import PostingProfilePIC from "./widget/PostingProfilePIC.vue";
import DimensionEditor from "@/components/common/DimensionEditor.vue";

layoutStore().name = "tenant";

const FEATUREID = "PostingProfile";
const profile = authStore().getRBAC(FEATUREID);

const listControl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: profile.canUpdate ? "edit" : "view",
  disableDirectPosting: false,
  disableDoNotPost: false,
});

function newRecord(record) {
  record._id = "";
  record.Name = "";
  record.Enable = true;
  record.IsActive = true;

  openForm(record);
}

function openForm() {
  util.nextTickN(2, () => {
    listControl.value.setFormFieldAttr("_id", "rules", [
      (v) => {
        let vLen = 0;
        let consistsInvalidChar = false;

        v.split("").forEach((ch) => {
          vLen++;
          const validCar =
            "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz-_".indexOf(
              ch
            ) >= 0;
          if (!validCar) consistsInvalidChar = true;
          //console.log(ch,vLen,validCar)
        });

        if (vLen < 3 || consistsInvalidChar)
          return "minimal length is 3 and alphabet only";
        return "";
      },
    ]);
  });
}
</script>
