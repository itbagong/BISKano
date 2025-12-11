<template>
  <div>
    <div class="flex flex-col gap-5 p-4">
      <div class="flex gap-1">
        <h1 class="grow">{{ title }}</h1>
        <div class="flex gap-1">
          <slot name="buttons" :preview="data.preview"></slot>
          <div class="min-w-[55px] h-[30px]" v-if="data.loadingPdf">
            <loader kind="skeleton" skeleton-kind="input" />
          </div>
          <SButton
            :disabled="disablePrint"
            v-else
            class="btn_primary"
            label="Print"
            icon="printer"
            @click="downloadPdf"
          ></SButton>
          <SButton
            class="btn_primary"
            label="Close"
            icon="Close"
            @click="emit('close')"
          ></SButton>
        </div>
      </div>
      <div class="flex flex-col gap-2" v-if="!data.loading">
        <slot name="header_preview_1" :preview="data.preview"> </slot>
        <slot name="header_preview" :preview="data.preview">
          <div v-if="!data.preview.HideHeader && data.preview.Header">
            <div
              v-for="(header, header_idx) in data.preview.Header.Data"
              :key="header_idx"
              class="gap-[1px] grid"
              :class="{
                gridCol1: header.length == 1,
                gridCol2: header.length == 2,
                gridCol3: header.length == 3,
                gridCol4: header.length == 4,
                gridCol5: header.length == 5,
                gridCol6: header.length == 6,
                gridCol7: header.length == 7,
                gridCol8: header.length == 8,
                gridCol9: header.length == 9,
                gridCol10: header.length == 10,
                gridCol11: header.length == 11,
                gridCol12: header.length == 12,
                gridCol13: header.length == 13,
                gridCol14: header.length == 14,
                gridCol15: header.length == 15,
                gridCol16: header.length == 16,
                gridCol17: header.length == 17,
              }"
            >
              <div v-for="(v, v_idx) in header" :key="v_idx">
                {{ v }}
              </div>
            </div>
          </div>
        </slot>
        <slot name="header_preview_2" :preview="data.preview"> </slot>
        <div
          v-for="(section, index) in data.preview.Sections"
          class="my-8"
          :key="index"
        >
          <h2 v-if="!section.HideTitle && !hideSectionTitle">
            {{ section.Title }}
          </h2>
          <div v-if="!section.HideHeader"></div>
          <div
            v-for="(item, idx) in section.Items"
            class="gap-[1px] even:bg-slate-100 grid"
            :key="'idx_' + idx"
            :class="{
              gridCol1: item.length == 1,
              gridCol2: item.length == 2,
              gridCol3: item.length == 3,
              gridCol4: item.length == 4,
              gridCol5: item.length == 5,
              gridCol6: item.length == 6,
              gridCol7: item.length == 7,
              gridCol8: item.length == 8,
              gridCol9: item.length == 9,
              gridCol10: item.length == 10,
              gridCol11: item.length == 11,
              gridCol12: item.length == 12,
              gridCol13: item.length == 13,
              gridCol14: item.length == 14,
              gridCol15: item.length == 15,
              gridCol16: item.length == 16,
              gridCol17: item.length == 17,
            }"
          >
            <div
              v-for="(v, idxCol) in item"
              :key="idxCol"
              :class="{
                gridHeader: idx == 0,
                'text-right':
                  data.objClass[index + '#' + idxCol] == 'R' && idx > 0,
                'text-center':
                  data.objClass[index + '#' + idxCol] == 'C' && idx > 0,
              }"
            >
              {{ idx == 0 ? formatHeader(index, idxCol, v) : v }}
            </div>
          </div>
        </div>

        <div v-if="!hideFooter && data.preview.Header?.Footer">
          <slot name="footer_preview" :preview="data.preview">
            <div
              v-if="data.preview.Header.Footer"
              v-for="(footer, footer_idx) in data.preview.Header.Footer"
              :key="footer_idx"
              class="gap-[1px] grid font-semibold"
              :class="{
                gridCol1: footer.length == 1,
                gridCol2: footer.length == 2,
                gridCol3: footer.length == 3,
                gridCol4: footer.length == 4,
                gridCol5: footer.length == 5,
                gridCol6: footer.length == 6,
                gridCol7: footer.length == 7,
                gridCol8: footer.length == 8,
                gridCol9: footer.length == 9,
                gridCol10: footer.length == 10,
                gridCol11: footer.length == 11,
                gridCol12: footer.length == 12,
                gridCol13: footer.length == 13,
                gridCol14: footer.length == 14,
                gridCol15: footer.length == 15,
                gridCol16: footer.length == 16,
                gridCol17: footer.length == 17,
              }"
            >
              <div
                v-for="(v, v_idx) in footer"
                :key="v_idx"
                :class="`${v.toLowerCase().replace(/[ 0-9:.,]/g, '')} ${
                  data.objClass['footer#' + v_idx] == 'R'
                    ? 'text-right'
                    : data.objClass['footer#' + v_idx] == 'C'
                    ? 'text-center'
                    : 'text-left'
                }`"
              >
                {{
                  v_idx == footer.length - 1
                    ? formatHeader("footer", v_idx, v)
                    : v
                }}
              </div>
            </div>
          </slot>
        </div>
        <div v-if="!hideFooter1">
          <slot name="footer_preview_1" :preview="data.preview"></slot>
        </div>
        <div v-if="!hideSignature">
          <div class="flex gap-5 pt-10 justify-center">
            <div class="flex gap-5">
              <template
                v-if="data.preview.Signature"
                v-for="(item, index) in data.preview.Signature"
                :key="index"
              >
                <div class="w-[300px]">
                  <h3 class="font-semibold mb-4 text-center">
                    {{ item.Header }},
                  </h3>
                  <div class="relative border-b border-gray-300 h-32 mb-4">
                    <div class="img-status-signature">
                      <img
                        v-if="item.Status === 'APPROVED'"
                        src="https://xibar-dev.s3.ap-southeast-3.amazonaws.com/template/images/status_approved.jpeg"
                        width="130"
                      />

                      <img
                        v-if="item.Status === 'REJECTED'"
                        src="https://xibar-dev.s3.ap-southeast-3.amazonaws.com/template/images/status_rejected.jpeg"
                        width="130"
                      />
                    </div>

                    <span class="absolute bottom-1 w-full text-center">{{
                      `${item.Footer}`
                    }}</span>
                  </div>
                  <div class="text-center">
                    <p class="text-sm">
                      {{
                        item.Confirmed
                          ? `Tgl. ${moment(item.Confirmed).format(
                              "DD-MMM-yyyy HH:mm:ss"
                            )}`
                          : ""
                      }}
                    </p>
                  </div>
                </div>
              </template>
            </div>
          </div>
          <!-- multi signature -->
          <template
            v-if="data.preview.MultipleRowSignature"
            v-for="(row, i) in data.preview.MultipleRowSignature"
            :key="i"
          >
            <div class="flex gap-5 pt-10 justify-center">
              <div class="flex gap-5 justify-between">
                <template v-for="(item, index) in row" :key="index">
                  <div class="w-[300px]">
                    <h3 class="font-semibold mb-4 text-center">
                      {{ item.Header }},
                    </h3>
                    <div class="relative border-b border-gray-300 h-32 mb-4">
                      <div class="img-status-signature">
                        <img
                          v-if="item.Status === 'APPROVED'"
                          src="https://xibar-dev.s3.ap-southeast-3.amazonaws.com/template/images/status_approved.jpeg"
                          width="130"
                        />

                        <img
                          v-if="item.Status === 'REJECTED'"
                          src="https://xibar-dev.s3.ap-southeast-3.amazonaws.com/template/images/status_rejected.jpeg"
                          width="130"
                        />
                      </div>

                      <span class="absolute bottom-1 w-full text-center">{{
                        `${item.Footer}`
                      }}</span>
                    </div>
                    <div class="text-center">
                      <p class="text-sm">
                        {{
                          item.Confirmed
                            ? `Tgl. ${moment(item.Confirmed).format(
                                "DD-MMM-yyyy HH:mm:ss"
                              )}`
                            : ""
                        }}
                      </p>
                    </div>
                  </div>
                </template>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>
    <div
      class="w-full text-center font-semibold"
      v-if="!data.loading && data.preview.Sections == nil"
    >
      No Data
    </div>
    <loader kind="skeleton" skeleton-kind="list" v-if="data.loading" />
  </div>
</template>

<script setup>
import { SButton, util } from "suimjs";
import moment from "moment";
import { reactive, ref, inject, onMounted } from "vue";
import Loader from "@/components/common/Loader.vue";

const props = defineProps({
  title: { type: String, default: () => "" },
  preview: { type: Object, default: () => {} },
  SourceType: { type: String, default: () => "" },
  SourceJournalID: { type: String, default: () => "" },
  Name: { type: String, default: () => "Default" },
  VoucherNo: { type: String, default: () => "" },
  hideFooter: { type: Boolean, default: false },
  hideFooter1: { type: Boolean, default: false },
  hideSignature: { type: Boolean, default: true },
  hideSectionTitle: { type: Boolean, default: false },
  reload: { type: Number, default: () => 1 },
  disablePrint: { type: Boolean },
});

const axios = inject("axios");

const emit = defineEmits({
  close: null,
  updatePrint: null,
});

const data = reactive({
  preview: {},
  objClass: {},
  loading: false,
  loadingPdf: false,
});

function loadPreview() {
  let url = "";
  console.log(props.SourceType);
  switch (props.SourceType) {
    case "WORKTERMINATION":
    case "CONTRACT":
    case "OVERTIME":
    case "MANPOWER":
    case "COACHING":
    case "PLOTTING":
    case "BUSINESSTRIP":
    case "LEAVECOMPENSATION":
    case "SK":
    case "ASSESSMENT":
    case "TALENTDEVELOPMENT":
      url = `/hcm/preview/get?reload=${props.reload}&type=${props.SourceType}&id=${props.SourceJournalID}&name=${props.Name}&voucher=${props.VoucherNo}`;
      break;
    case "Inventory Receive":
    case "Inventory Issuance":
    case "Movement In":
    case "Movement Out":
    case "Transfer":
    case "Asset Acquisition":
    case "Item Request":
    case "Purchase Request":
    case "INVENTORY":
    case "Purchase Order":
      url = `/scm/preview/get?reload=${props.reload}&type=${props.SourceType}&id=${props.SourceJournalID}&name=${props.Name}&voucher=${props.VoucherNo}`;
      break;
    case "Work Request":
    case "Work Order":
    case "Work Order Report Consumption":
    case "Work Order Report Resource":
    case "Work Order Report Output":
      url = `/mfg/preview/get?reload=1&type=${props.SourceType}&id=${props.SourceJournalID}&name=${props.Name}&voucher=${props.VoucherNo}`;
      break;
    case "CONTRACT":
    case "OVERTIME":
    case "LOAN":
    case "WORKTERMINATION":
    case "MANPOWER":
    case "COACHING":
    case "PLOTTING":
    case "BUSINESSTRIP":
    case "LEAVECOMPENSATION":
    case "SK":
    case "ASSESSMENT":
    case "TALENTDEVELOPMENT":
    case "TRAININGDEVELOPMENT":
    case "TRAININGDEVELOPMENTDETAIL":
      url = `/hcm/preview/get?reload=${props.reload}&type=${props.SourceType}&id=${props.SourceJournalID}&name=${props.Name}&voucher=${props.VoucherNo}`;
      break;
    default:
      url = `/fico/preview/get?reload=${props.reload}&type=${props.SourceType}&id=${props.SourceJournalID}&name=${props.Name}&voucher=${props.VoucherNo}`;
  }
  data.loading = true;
  axios.post(url).then(
    (r) => {
      data.preview = r.data;
      data.loading = false;
    },
    (e) => {
      data.loading = false;
      util.showError(e);
    }
  );
}
function downloadPdf() {
  data.loadingPdf = true;
  const url = `/fico/postingprofile/preview-download-as-pdf?SourceType=${props.SourceType}&SourceJournalID=${props.SourceJournalID}`;
  axios
    .get(url, {
      responseType: "blob",
    })
    .then(
      (r) => {
        const downloadUrl = window.URL.createObjectURL(r.data);
        window.open(downloadUrl, "__blank");
        window.URL.revokeObjectURL(url);
        data.loadingPdf = false;
        emit("updatePrint", props.SourceType, props.SourceJournalID);
      },
      (e) => {
        data.loadingPdf = false;
        util.showError(e);
      }
    );
}
function formatHeader(idx, col, v) {
  let splited = v.split(":");
  if (splited.length > 1) {
    data.objClass[idx + "#" + col] = splited[1];
  }
  return splited.length > 1 ? splited[0] : v;
}

onMounted(() => {
  loadPreview();
});
</script>
<style scoped>
.gridCol1 {
  @apply grid-cols-1;
}
.gridCol2 {
  @apply grid-cols-2;
}
.gridCol3 {
  @apply grid-cols-3;
}
.gridCol4 {
  @apply grid-cols-4;
}
.gridCol5 {
  @apply grid-cols-5;
}
.gridCol6 {
  @apply grid-cols-6;
}
.gridCol7 {
  @apply grid-cols-7;
}
.gridCol8 {
  @apply grid-cols-8;
}
.gridCol9 {
  @apply grid-cols-9;
}
.gridCol10 {
  @apply grid-cols-10;
}
.gridCol11 {
  @apply grid-cols-11;
}
.gridCol12 {
  @apply grid-cols-12;
}
.gridCol13 {
  @apply grid-cols-13;
}
.gridCol14 {
  @apply grid-cols-14;
}
.gridCol15 {
  @apply grid-cols-15;
}
.gridCol16 {
  @apply grid-cols-16;
}
.gridCol17 {
  @apply grid-cols-17;
}
.gridHeader {
  @apply font-bold border-b bg-primary p-1 text-white border-slate-400;
}
.text-R {
  @apply text-right;
}
.text-C {
  @apply text-center;
}
.img-status-signature {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  padding-bottom: 36px;
}
</style>
