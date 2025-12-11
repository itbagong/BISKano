<template>
  <div class="w-full bg-white relative upload-master">
    <div
      class="absolute w-full h-full bg-[#dddddda6] z-[15]"
      v-if="data.loadingDownload || data.loadingUpload"
    >
      <div class="flex justify-center items-center h-full w-full text-center">
        <div class="flex-col gap-4 w-[300px]">
          <div class="font-semibold flex justify-center items-center mb-1">
            <template v-if="data.loadingDownload">
              <mdicon name="download" size="22" class="mr-[2px]" />
              <h1 class="">Download...</h1>
            </template>
            <template v-else>
              <mdicon name="Upload" size="22" class="mr-[2px]" />
              <h1>Upload...</h1>
            </template>
          </div>
          <loader class="loader-linier" />
        </div>
      </div>
    </div>
    <div class="p-2">
      <div></div>
      <div class="flex gap-8 items-center mb-4 pb-4 border-b-[1px]">
        <s-input
          label="Master Name"
          use-list
          class="w-[400px]"
          :items="data.masterList"
          v-model="data.masterSelected"
          @change="onChangeMaster"
        />

        <div class="flex gap-4 mt-3">
          <input
            type="file"
            hidden
            ref="fileRef"
            @change="onPickFile"
            accept=".xlsx, .xls"
          />
          <s-button
            icon="download"
            :disabled="data.masterSelected == null"
            label="Download Template"
            @click="onDownloadFile"
            class="btn_primary"
          />
          <s-button
            icon="upload"
            :disabled="data.masterSelected == null"
            label="Upload"
            @click="onUploadFile"
            class="btn_primary"
          />
        </div>
      </div>
      <template v-if="data.gridCfg.setting">
        <div class="flex gap-8 items-center justify-between">
          <h1 class="mb-3">Data</h1>
          <s-button
            v-if="data.serviceDownload"
            icon="download"
            label="Download Data"
            @click="downloadData"
            class="btn_primary"
          />
        </div>

        <s-grid
          :key="data.serviceSelected"
          ref="gridCtl"
          :config="data.gridCfg"
          :readUrl="data.serviceSelected + '/gets'"
          hide-select
          hide-refresh-button
          hide-new-button
          hide-action
        />
      </template>
    </div>
  </div>
</template>
<script setup>
import { reactive, ref, computed, inject, watch, onMounted } from "vue";
import { layoutStore } from "@/stores/layout.js";
import { util, SButton, SInput, SModal, SGrid, loadGridConfig } from "suimjs";
import Loader from "@/components/common/Loader.vue";
import Helper from "@/scripts/helper.js";

layoutStore().name = "tenant";

const axios = inject("axios");
const gridCtl = ref(null);
const fileRef = ref(null);
const data = reactive({
  masterList: [],
  masterSelected: null,
  serviceSelected: "",
  serviceDownload: "",
  gridCfg: {},
  gridKey: 0,
  loadingUpload: false,
  loadingDownload: false,
});
function uploadFile(content) {
  data.loadingUpload = true;

  axios
    .post("/bagong/masterupload/upload", {
      TableName: data.masterSelected,
      Content: content,
    })
    .then(
      (r) => {
        downloadFile(r.data.Body, r.data.Header);
      },
      (e) => util.showError(e)
    )
    .finally(() => {
      data.loadingUpload = false;
    });
}
function onUploadFile() {
  fileRef.value.click();
}
function onPickFile() {
  const handler = fileRef.value;
  if (handler.files.length == 0) return;
  const fs = handler.files[0];
  const reader = new FileReader();
  reader.onload = function () {
    const binaryString = reader.result;
    uploadFile(binaryString.split(",")[1]);
  };
  reader.readAsDataURL(fs);
}
function downloadFile(body, header) {
  var binaryString = window.atob(body);
  var binaryLen = binaryString.length;
  var bytes = new Uint8Array(binaryLen);
  for (var i = 0; i < binaryLen; i++) {
    var ascii = binaryString.charCodeAt(i);
    bytes[i] = ascii;
  }

  var blob = new Blob([bytes], {
    type: header["Content-Type"]?.join(""),
  });

  var fileName = "";
  var disposition = header["Content-Disposition"]?.join("");
  if (disposition && disposition.indexOf("attachment") !== -1) {
    var filenameRegex = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/;
    var matches = filenameRegex.exec(disposition);
    if (matches != null && matches[1])
      fileName = matches[1].replace(/['"]/g, "");
  }

  var link = document.createElement("a");
  link.href = window.URL.createObjectURL(blob);
  link.download = fileName;
  link.click();

  setTimeout(() => {
    link.remove();
  }, 1500);
}
function onDownloadFile() {
  data.loadingDownload = true;

  axios
    .post("/bagong/masterupload/download-template-excel", {
      TableName: data.masterSelected,
      ExcelFile: data.masterSelected,
      IsTemplate: true,
    })
    .then(
      (r) => {
        downloadFile(r.data.Body, r.data.Header);
      },
      (e) => util.showError(e)
    )
    .finally(() => {
      data.loadingDownload = false;
    });
}
function fetchMaster() {
  axios.post("/bagong/masterupload/get-list-models", {}).then(
    (r) => {
      data.masterList = r.data.map((e) => {
        return {
          ...e,
          key: e.Key,
          text: e.Name,
        };
      });
    },
    (e) => util.showError(e)
  );
}
function initGridConfig() {
  loadGridConfig(axios, data.serviceSelected + "/gridconfig").then(
    (r) => {
      data.gridCfg = r;
    },
    (e) => {
      data.gridCfg = {};
      util.showError(e);
    }
  );
}
function onChangeMaster(v1, v2, item) {
  const el = data.masterList.find((e) => v2 == e.Key) ?? { Service: "" };
  data.serviceSelected = el.Service;
  data.serviceDownload = el.Download;
  if (data.serviceSelected == "") {
    data.gridCfg = {};
  } else {
    initGridConfig();
  }
}

function downloadData(content) {
  data.loadingDownload = true;

  axios
    .post("/bagong/masterupload/download", {
      TableName: data.masterSelected,
      Content: content,
    })
    .then(
      (r) => {
        downloadFile(r.data.Body, r.data.Header);
      },
      (e) => util.showError(e)
    )
    .finally(() => {
      data.loadingDownload = false;
    });
}

onMounted(() => {
  fetchMaster();
});
</script>
<style scoped>
.upload-master .loader-linier {
  height: 12px;
  border-radius: 4px;
}
</style>
