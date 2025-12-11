<template>
  <div>
    <data-list
      class="card"
      ref="listControl"
      title="Item Template"
      grid-config="/tenant/itemtemplate/gridconfig"
      form-config="/tenant/itemtemplate/formconfig"
      grid-read="/tenant/itemtemplate/gets"
      form-read="/tenant/itemtemplate/get"
      grid-mode="grid"
      grid-delete="/tenant/itemtemplate/delete"
      form-keep-label
      form-insert="/tenant/itemtemplate/save"
      form-update="/tenant/itemtemplate/save"
      :grid-fields="['Enable']"
      :form-fields="['Dimension']"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :form-tabs-new="data.tablist"
      :form-tabs-edit="data.tablist"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @preSave="onPreSave"
    >
      <template #form_input_Dimension="{ item }">
        <dimension-editor-vertical
          v-model="item.Dimension"
        ></dimension-editor-vertical>
      </template>

      <template #form_tab_Detail>
        <parent-child v-model="data.detailItems" :max-level="3" />
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, computed, inject } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput, SButton } from "suimjs";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import ParentChild from "./widget/ItemTemplateChild.vue";

layoutStore().name = "tenant";

const axios = inject("axios");

const listControl = ref(null);

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  record: {},
  tablist: ["General", "Detail"],
  detailItems: [],
  idxParent: 0,
});

function newRecord(record) {
  record._id = "";
  record.Module = "";
  record.Menu = "";
  record.Name = "";
  record.Status = false;
  record.IsActive = true;
  data.detailItems = [];
  openForm(record);
}
function editRecord(record) {
  data.formMode = "edit";
  data.record = record;
  data.detailItems = buildHierarchy(record.Items);
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

function onPreSave(record) {
  record.Items = unWind(data.detailItems);
}

function unWind(arry) {
  let res = [];

  let unwindChild = function (dt) {
    for (let i in dt) {
      let ob = dt[i];
      res.push(ob);
      if (ob.Child.length > 0) unwindChild(ob.Child);
    }
  };

  for (let i in arry) {
    let ob = arry[i];
    res.push(ob);
    if (ob.Child.length > 0) unwindChild(ob.Child);
  }

  return res;
}

function buildHierarchy(arry) {
  let res = [],
    objArr = {};
  for (let i in arry) {
    let ob = arry[i],
      p = ob.Parent;
    ob.Child = [];
    ob.level = ob.ID.split("#$").length > 0 ? ob.ID.split("#$").length : 1;
    objArr[ob.ID] = ob;
  }

  for (let ky in objArr) {
    if (objArr[ky].Parent !== "") {
      objArr[objArr[ky].Parent].Child.push(objArr[ky]);
    } else {
      res.push(objArr[ky]);
    }
  }

  return res;
}
</script>
