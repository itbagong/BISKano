<template>
  <div class="flex flex-col gap-2">
    <s-grid
      ref="listControl"
      class="w-full tb-attachment"
      editor
      hide-search
      hide-sort
      :hide-new-button="false"
      hide-refresh-button
      :hide-detail="true"
      hide-select
      auto-commit-line
      no-confirm-delete
      :config="data.gridCfg"
      form-keep-label
      @new-data="newRecord"
      @delete-data="onDelete"
    >
      <template #item_File="{ item }">
        <s-input v-model="item.File.URI" hidden></s-input>
        <input
          v-if="item.File.FileName == ''"
          type="file"
          @change="
            (event) => {
              handleFileUpload(event, item);
            }
          "
        />
        <div v-else>{{ item.File.FileName }}</div>
      </template>
    </s-grid>
  </div>
</template>
<script setup>
import { onMounted, inject, computed, watch, reactive, ref } from "vue";
import { loadGridConfig, util, SInput, SGrid } from "suimjs";
const axios = inject("axios");
const refItemID = ref(null);
const refSKU = ref(null);
const refUom = ref(null);
const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  item: { type: Object, default: () => {} },
  gridConfig: { type: String, default: () => "" },
  gridRead: { type: String, default: () => "" },
  readOnly: { type: Boolean, defaule: false },
  hideDetail: { type: Boolean, defaule: false },
});

const emit = defineEmits({
  "update:modelValue": null,
  recalc: null,
});

const listControl = ref(null);

const data = reactive({
  records:
    props.modelValue == null || props.modelValue == undefined
      ? []
      : props.modelValue,
  typeMoveIn: "line",
  gridCfg: {},
});

function newRecord() {
  const record = {};
  record._id = util.uuid();
  record.WorkRequestID = props.item._id;
  record.Description = "";
  record.File = {
    URI: "",
    FileName: "",
    ContentType: "",
    Size: "",
  };
  listControl.value.setRecords([...listControl.value.getRecords(), record]);
}

function onDelete(record, index) {
  const newRecords = listControl.value.getRecords().filter((dt, idx) => {
    return idx != index;
  });
  listControl.value.setRecords(newRecords);
}

function handleFileUpload(event, item) {
  const file = event.target.files[0];
  const fileName = file.name;
  const fileURI = URL.createObjectURL(file);
  const fileType = file.type;
  const fileSize = file.size;
  let Attachment = listControl.value.getRecords();
  Attachment.map(function (v) {
    if (v._id == item._id) {
      let file = {
        FileName: fileName,
        URI: fileURI,
        ContentType: fileType,
        Size: fileSize,
      };
      v.File = file;
    }
    return v;
  });
  listControl.value.setRecords(Attachment);
}

function getDataValue() {
  return listControl.value.getRecords();
}

function genCfgAttachment() {
  let Attachment = [
    {
      field: "Description",
      kind: "Text",
      label: "Description",
      readType: "show",
      input: {
        field: "Description",
        label: "Description",
        hint: "",
        hide: false,
        placeHolder: "Description",
        kind: "text",
        disable: false,
        required: false,
        multiple: false,
      },
    },
    {
      field: "File",
      kind: "Text",
      label: "File",
      readType: "show",
      input: {
        field: "File",
        label: "File",
        hint: "",
        hide: false,
        placeHolder: "File",
        kind: "text",
        disable: false,
        required: false,
        multiple: false,
      },
    },
  ];
  data.gridCfg = {
    fields: Attachment,
  };
}

onMounted(() => {
  genCfgAttachment();
  axios
    .post(props.gridRead, {
      Skip: 0,
      Take: 0,
      Sort: ["_id"],
    })
    .then(
      (r) => {
        if (listControl.value) {
          listControl.value.setRecords(r.data.data);
        }
      },
      (e) => util.showError(e)
    );
});
defineExpose({
  getDataValue,
});
</script>
