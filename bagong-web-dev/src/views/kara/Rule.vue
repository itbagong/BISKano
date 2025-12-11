<template>
  <div class="w-full">
    <data-list
      class="shadow-md border rounded-md"
      ref="listControl"
      v-if="data.gridConfig != ''"
      :title="data.title"
      :grid-config="data.gridConfig"
      :grid-read="data.gridRead"
      :grid-delete="data.gridDelete"
      :form-config="data.formConfig"
      :form-read="data.formRead"
      grid-mode="grid"
      :form-insert="data.formInsert"
      :form-update="data.formUpdate"
      :grid-fields="['Enable']"
      :form-tabs-edit="['General', 'Lines']"
      stay-on-form-after-save
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      @formNewData="newRecord"
      @formEditData="openForm"
      :grid-custom-filter="data.customFilter"
      @gridResetCustomFilter="resetGridHeaderFilter"
    >
      <template #grid_header_search>
        <grid-header-filter
          ref="gridHeaderFilter"
          v-model="data.customFilter"
          hide-filter-text
          hide-filter-date
          hide-filter-status
          @init-new-item="initNewItemFilter"
          @pre-change="changeFilter"
          @change="refreshGrid"
        >
          <template #filter_1="{ item }">
            <s-input
              class="w-[200px]"
              keep-label
              label="Search Name"
              v-model="item.Keyword"
            />
            <s-input
              label="Period start"
              kind="date"
              v-model="item.PeriodStart"
            />
            <s-input
              label="Period end"
              kind="date"
              v-model="item.PeriodEnd"
            />
          </template>
        </grid-header-filter>
      </template>
      <template #form_tab_Lines="{ item }">
        <rule-line :rule="item" />
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, watch } from "vue";
import { layoutStore } from "@/stores/layout.js";

import { authStore } from "@/stores/auth.js";
import { DataList, util, SInput } from "suimjs";
import { useRoute } from "vue-router";
import RuleLine from "./widget/RuleLine.vue";
import GridHeaderFilter from "@/components/common/GridHeaderFilter.vue";

layoutStore().name = "tenant";

const FEATUREID = "AttendanceRule";
const profile = authStore().getRBAC(FEATUREID);

const listControl = ref(null);
const gridHeaderFilter = ref(null);

const route = useRoute();

const data = reactive({
  appMode: "grid",
  formMode: "edit",
  title: "Attendance Rules",
  gridConfig: "/kara/rule/gridconfig",
  gridRead: "/kara/rule/gets",
  gridDelete: "/kara/admin/rule/delete",
  formConfig: "/kara/rule/formconfig",
  formRead: "/kara/rule/get",
  formInsert: "/kara/admin/rule/insert",
  formUpdate: "/kara/admin/rule/update",
  customFilter: null,
});

function newRecord(record) {
  record._id = "";
  record.Name = "";
  record.Enable = true;

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

        if (vLen < 5 || consistsInvalidChar)
          return "minimal length is 5 and alphabet only";
        return "";
      },
    ]);
  });
}

function initNewItemFilter(item) {
  item.Keyword = "";
  item.PeriodStart = null;
  item.PeriodEnd = null;
}

function changeFilter(item, filters) {
  if (item.Keyword && item.Keyword != "") {
    filters.push({
      Op: "$contains",
      Field: "Name",
      Value: [item.Keyword],
    });
  }
  if (item.PeriodStart != null) {
    filters.push({
      Op: "$gte",
      Field: "PeriodStart",
      Value: new Date(item.PeriodStart),
    });
  }
  if (item.PeriodEnd != null) {
    filters.push({
      Op: "$lte",
      Field: "PeriodEnd",
      Value: new Date(item.PeriodEnd),
    });
  }
}

function refreshGrid() {
  listControl.value.refreshGrid();
}

function resetGridHeaderFilter() {
  gridHeaderFilter.value.reset();
}
</script>
