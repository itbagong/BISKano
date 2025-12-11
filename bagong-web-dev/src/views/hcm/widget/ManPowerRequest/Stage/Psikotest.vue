<template>
  <Loader kind="skeleton" v-if="data.loading" />
  <div v-else>
    <status-text :txt="data.record.Status" />
    <s-form
      ref="frmCtl"
      v-model="data.record"
      :config="data.frmCfg"
      keep-label
      hide-cancel
      :buttons-on-bottom="false"
      buttons-on-top
      @submit-form="submitForm"
    >
      <template #input_Details="{ item }">
        <div class="flex tab_container grow">
          <div
            v-for="(r, idx) in item.Details"
            :key="idx"
            :class="data.selectedTab === r.TemplateID ? 'tab_selected' : 'tab'"
            @click="data.selectedTab = r.TemplateID"
          >
            <s-input
              v-model="r.TemplateID"
              read-only
              class="w-full mb-2"
              use-list
              lookup-url="/she/mcuitemtemplate/find"
              lookup-key="_id"
              :lookup-labels="['Name']"
            />
          </div>
        </div>
        <div class="mb-5"></div>
        <div
          class="grid grid-cols-1 gap-y-2 shadow-lg p-2 border-t-8 border-primary mb-4"
        >
          <div v-if="selectedItem.Detail">
            <div v-if="!Array.isArray(selectedItem.Detail)">
              <div class="text-center border py-2 mb-2">
                <p class="text-md font-medium mb-1">IQScore</p>
                <p class="text-4xl font-medium">
                  {{ selectedItem.Detail?.IQScore }}
                </p>
              </div>
              <div class="text-center border py-2">
                <p class="text-md font-medium mb-1">Score</p>
                <p class="text-4xl font-medium">
                  {{ selectedItem.Detail?.Score }}
                </p>
              </div>
            </div>
            <div
              v-else
              class="w-full grid grid-cols-2 border-l border-t border-b"
            >
              <div
                v-for="(r, idx) in selectedItem.Detail"
                :key="idx"
                class="w-full border-r"
              >
                <p
                  :class="`w-full text-center font-medium border-b uppercase ${
                    r.IsMostQuestionType && 'bg-slate-100'
                  }`"
                >
                  {{
                    r.IsMostQuestionType
                      ? "Menggambarkan Diri"
                      : "Tidak Menggambarkan Diri"
                  }}
                </p>
                <div
                  v-for="(pr, idx) in r.Detail"
                  :key="idx"
                  :class="`grid grid-cols-2 border-b ${
                    r.IsMostQuestionType && 'bg-slate-100'
                  }`"
                >
                  <p class="text-center border-r">{{ pr.Answer }}</p>
                  <p class="text-center">{{ pr.Count }}</p>
                </div>
              </div>
            </div>
          </div>
          <div v-else class="border p-2">
            <p class="text-center">No Data Score</p>
          </div>
        </div>
      </template>
    </s-form>
  </div>
</template>
<script setup>
import { reactive, ref, inject, onMounted, watch, computed } from "vue";
import { util, SForm, SInput, loadFormConfig } from "suimjs";
import StatusText from "@/components/common/StatusText.vue";
import Loader from "@/components/common/Loader.vue";

const props = defineProps({
  id: { type: String, default: "" },
  item: { type: Object, default: null },
});

const axios = inject("axios");
const frmCtl = ref(null);

const data = reactive({
  frmCfg: {},
  record: {},
  loading: "",
  selectedTab: "",
});

const selectedItem = computed(() => {
  return data.record.Details.find((v) => v.TemplateID === data.selectedTab);
});

function fetchRecord() {
  data.loading = true;
  axios
    .post("/hcm/psychologicaltest/get", [props.id])
    .then((r) => {
      data.record = props.item;
      data.selectedTab = props.item?.Details?.[0]?.TemplateID;
    })
    .catch((e) => {
      util.showError(e);
    })
    .finally(() => {
      data.loading = false;
    });
}

function submitForm(record, cbOk, cbError) {
  axios
    .post("/hcm/psychologicaltest/update", record)
    .then((r) => {
      cbOk();
    })
    .catch((e) => {
      cbError();
      util.showError(e);
    });
}

onMounted(() => {
  loadFormConfig(axios, "/hcm/psychologicaltest/formconfig").then(
    (r) => {
      data.frmCfg = r;
      fetchRecord();
    },
    (e) => util.showError(e)
  );
});

watch(
  () => props.id,
  () => {
    fetchRecord();
  }
);
</script>
