<template>
  <div>
    <data-list
      ref="listControl"
      hide-title
      no-gap
      grid-editor
      grid-hide-search
      grid-hide-sort
      grid-hide-refresh
      grid-hide-detail
      grid-hide-select
      grid-no-confirm-delete
      init-app-mode="grid"
      grid-mode="grid"
      new-record-type="grid"
      :grid-config="gridConfig"
      :grid-fields="gridFields"
      grid-auto-commit-line
      @grid-row-add="newRecord"
      @grid-row-delete="onGridRowDelete"
      @grid-row-field-changed="onGridRowFieldChanged"
    >
      <template #grid_FileName="{ item }">
        <s-input v-model="item.FileName" hidden></s-input>
        <div>{{ item.FileName.replace(/\.[^/.]+$/, "") }}</div>
      </template>
      <template #grid_UploadDate="{ item }">
        <s-input v-model="item.UploadDate" hidden></s-input>
        <div>{{ moment(item.UploadDate).format("DD MMMM YYYY") }}</div>
      </template>
      <template #grid_URI="{ item }">
        <s-input v-model="item.URI" hidden></s-input>
        <input
          v-if="item.FileName == ''"
          type="file"
          @change="
            (event) => {
              handleFileUpload(event, item);
            }
          "
        />
        <div v-else>{{ item.FileName }}</div>
      </template>
    </data-list>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted, watch } from "vue";
import { DataList, SInput } from "suimjs";
import moment from "moment";

const props = defineProps({
  siteEntryAssetID: { type: String, default: "" },
  gridConfig: { type: String, default: "" },
  gridFields: { type: Array, default: () => [] },
  modelValue: { type: Array, default: () => [] },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const data = reactive({
  records:
    props.modelValue == null || props.modelValue == undefined
      ? []
      : props.modelValue,
  fileRecords: [],
});

const listControl = ref(null);

function newRecord(r) {
  r.AttachID = genRandomId();
  r.FileName = "";
  r.Description = "";
  r.PIC = "";
  r.UploadDate = new Date().toISOString();

  data.records.push(r);
  updateItems();
}

function genRandomId() {
  const chars = "0123456789abcdef";
  const idLen = 24;
  let result = "";
  for (let i = 0; i < idLen; i++) {
    const randomIndex = Math.floor(Math.random() * chars.length);
    result += chars.charAt(randomIndex);
  }
  return result;
}

function onGridRowDelete(_, index) {
  const newRecords = data.records.filter((dt, idx) => {
    return idx != index;
  });
  data.records = newRecords;
  listControl.value.setGridRecords(data.records);
  updateItems();
}

function onGridRowFieldChanged(name, v1, v2, old, record) {
  listControl.value.setGridRecord(
    record,
    listControl.value.getGridCurrentIndex()
  );
  updateItems();
}

function handleFileUpload(event, item) {
  const file = event.target.files[0];
  const fileName = file.name;
  const fileURI = URL.createObjectURL(file);
  const fileType = file.type;
  const fileSize = file.size;
  let Attachment = listControl.value.getGridRecords();
  Attachment.map(function (v) {
    if (v.AttachID == item.AttachID) {
      v.FileName = fileName;
      v.URI = fileURI;
      v.ContentType = fileType;
      v.Size = fileSize;
    }
    return v;
  });
  listControl.value.setGridRecords(Attachment);
}

function updateItems() {
  listControl.value.setGridRecords(data.records);
  emit("update:modelValue", data.records);
}

onMounted(() => {
  setTimeout(() => {
    updateItems();
  }, 500);
});

watch(
  () => data.records,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
</script>
