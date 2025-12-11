<template>
  <div class="flex flex-col gap-2">
    <s-grid
      ref="listControl"
      class="w-full tb-material grid-line-items"
      v-model="data.value"
      hide-search
      hide-sort
      auto-commit-line
      no-confirm-delete
      :hide-refresh-button="true"
      :hide-new-button="true"
      :hide-detail="true"
      :hide-select="true"
      :hide-action="true"
      hide-paging
      :config="data.gridCfg"
      total-url="/mfg/workorderplan/gets-material-line"
      form-keep-label
    >
      <template #header_search="{ config }">
        <div class="w-full flex gap-2 justify-left items-center header">
          <s-input
            label="Date From"
            kind="date"
            class="w-[200px]"
            v-model="data.search.DateFrom"
          ></s-input>
          <s-input
            label="Date To"
            kind="date"
            class="w-[200px]"
            v-model="data.search.DateTo"
          ></s-input>
          <s-input
            v-if="props.type == 'Material'"
            ref="refItemIDs"
            label="Items"
            v-model="data.search.Items"
            use-list
            lookup-key="_id"
            lookup-url="/tenant/item/find"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
            class="w-full"
            :multiple="true"
          ></s-input>
        </div>
      </template>
      <template #header_buttons="{ config }">
        <s-button
          icon="refresh"
          class="btn_primary refresh_btn"
          @click="getsSummary"
        />
      </template>
      <template #grid_total="{ item }">
        <!-- <tr v-for="(dt, idx) in item" :key="idx" class="font-semibold">
          <td colspan="6" class="ml-4">Total</td>
          <td class="text-right">100</td>
          <td class="text-right">100</td>
          <td class="text-right">200</td>
          <td></td>
        </tr> -->
        <tr class="font-semibold">
          <td colspan="6">Total</td>
          <td class="text-right">{{ util.formatMoney(data.total, {}) }}</td>
          <td class="text-right"></td>
        </tr>
      </template>
    </s-grid>
  </div>
</template>
<script setup>
import { onMounted, inject, reactive, ref } from "vue";
import { loadGridConfig, util, SGrid, SInput, SButton } from "suimjs";
const axios = inject("axios");
const props = defineProps({
  modelValue: { type: [Object, Array], default: () => [] },
  item: { type: Object, default: () => {} },
  gridConfig: { type: String, default: () => "" },
  gridRead: { type: String, default: () => "" },
  type: { type: String, default: () => "Material" },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});

const listControl = ref(null);

const data = reactive({
  value: props.modelValue,
  search: { DateFrom: null, DateTo: null, Items: [] },
  gridCfg: {},
  total: 0,
});
function getsSummary() {
  const search = JSON.parse(JSON.stringify(data.search));
  if (props.item.Status == "") {
    return true;
  }
  const payload = {
    ...search,
    WorkOrderPlanID: props.item._id,
    Type: props.type,
  };
  payload.DateFrom =
    payload.DateFrom == "Invalid date"
      ? null
      : payload.DateFrom
      ? payload.DateFrom
      : null;
  payload.DateTo =
    payload.DateTo == "Invalid date"
      ? null
      : payload.DateTo
      ? payload.DateTo
      : null;
  listControl.value.setLoading(true);
  axios.post(`/mfg/workorderplan/gets-material-line`, payload).then(
    (r) => {
      util.nextTickN(2, () => {
        data.value = r.data.data;
        data.total = r.data.total;
        listControl.value.setLoading(false);
      });
    },
    (e) => {
      data.value = props.modelValue;
      listControl.value.setLoading(false);
      util.showError(e);
    }
  );
}

onMounted(() => {
  loadGridConfig(axios, props.gridConfig).then(
    (r) => {
      r.fields.map((f) => {
        if (props.type == "Material" && f.field == "QtyAvailable") {
          f.readType = "hide";
        }
        return f;
      });
      data.gridCfg = r;
      util.nextTickN(2, () => {
        getsSummary();
      });
    },
    (e) => util.showError(e)
  );
});
defineExpose({
  getsSummary,
});
</script>
<style scoped>
.title-header {
  font-size: 14px;
  font-weight: 600;
}
</style>
