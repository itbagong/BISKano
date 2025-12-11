<template>
  <s-card class="w-full suim_datalist widget-item-pica" hide-footer no-gap>
    <div class="grow" v-if="!hideTitle">
      <div class="title section_title font-bold border-b-[1px] border-y-black">
        PICA
      </div>
    </div>
    <data-list
      ref="listControl"
      hide-title
      no-gap
      :form-config="'/she/pica/itempica/formconfig'"
      init-app-mode="form"
      @alter-form-config="onAlterConfig"
      form-hide-submit
      form-hide-cancel
      form-keep-label
      :form-fields="['_id', 'SourceNumber', 'Status', 'EmployeeID']"
    >
      <template #form_input_SourceNumber="{ item, config }">
        <label class="flex input_label">
          {{ config.label }}
        </label>
        <div v-if="item.SourceNumber">{{ item.SourceNumber }}</div>
        <div v-else>-</div>
      </template>
      <template #form_input__id="{ item, config }">
        <label class="flex input_label">
          {{ config.label }}
        </label>
        <div v-if="item._id">{{ item._id }}</div>
        <div v-else>-</div>
      </template>
      <template #form_input_Status="{ item, config }">
        <label class="flex input_label">
          {{ config.label }}
        </label>
        <div v-if="item.Status"><status-text :txt="item.Status" /></div>
        <div v-else>-</div>
      </template>
      <template #form_input_EmployeeID="{ item }">
        <s-input
          :read-only="readOnly"
          label="Responsible Person (PIC)"
          v-model="item.EmployeeID"
          use-list
          lookup-url="/tenant/employee/find"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          :key="item.EmployeeID"
        />
      </template>
    </data-list>
  </s-card>
</template>
<script setup>
import {
  SInput,
  util,
  SButton,
  SCard,
  DataList,
  SGrid,
  loadGridConfig,
} from "suimjs";
import {
  reactive,
  ref,
  inject,
  watch,
  onMounted,
  computed,
  nextTick,
} from "vue";
import moment from "moment";
import StatusText from "./StatusText.vue";

const axios = inject("axios");
const listControl = ref(null);

const props = defineProps({
  readOnly: { type: Boolean, default: false },
  modelValue: { type: Object, default: () => {} },
  hideTitle: { type: Boolean, default: false },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const data = reactive({
  isEdit: false,
  record: {},
});

const objStatus = {
  Open: "rounded px-2 py-1 border-red-500 bg-red-200 text-red-900",
  Close: "rounded px-2 py-1 border-green-500 bg-green-200 text-green-900",
};

watch(
  () => data.record,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);

function onAlterConfig(config) {
  config.sectionGroups.map((sg) => {
    sg.sections.map((s) => {
      if (["Info", "Dimension", "TakeAction"].includes(s.title)) {
        s.visible = false;
      }
      return s;
    });
    return sg;
  });
  updateItems();
}

function updateItems() {
  util.nextTickN(2, () => {
    data.record = props.modelValue;
    listControl.value.setFormRecord(data.record);
    if (props.readOnly) listControl.value.setFormMode("view");
  });
}
</script>
<style>
.widget-item-pica .bg-white.suim_datalist {
  @apply bg-transparent;
}
</style>
