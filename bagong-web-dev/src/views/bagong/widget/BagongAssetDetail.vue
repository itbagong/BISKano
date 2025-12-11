<template>
  <div>
    <s-form
      :class="['s-form-assetdetail', assetMap[props.groupID]]"
      v-if="!data.loading"
      ref="frmCtl"
      v-model="data.record[assetMap[props.groupID]]"
      :config="data.config"
      keep-label
      only-icon-top
      hide-submit
      hide-cancel
      class="form_detail_asset_bagong"
      @fieldChange="fieldChange"
    >
      <template #input_RegisterInfo="{ config, item }">
        <RegisterInfo v-model="item.RegisterInfo" @close="emit('closeAttch')" :recordId="recordId"  />
      </template>
    </s-form>
    <!-- <s-form
      v-if="!data.loadingOther"
      v-model="data.record[assetMap[props.groupID]].OtherInfo"
      :config="data.formCfgOtherInfo"
      keep-label
      only-icon-top
      hide-submit
      hide-cancel
      class="form_detail_asset_bagong"
    >
    </s-form>-->
  </div>
</template>

<script setup>
import { reactive, onMounted, inject, watch, ref } from "vue";
import { SForm, SInput, loadFormConfig, util } from "suimjs";
import RegisterInfo from "./BagongAssetRegisterInfo.vue";

const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  recordId: { type: String, default: "" },
  readOnly: { type: Boolean, default: false },
  groupID: { type: String, default: "" },
  dimension: { type: Array, default: [] },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const assetMap = {
  ELC: "DetailElectronic",
  PRT: "DetailProperty",
  UNT: "DetailUnit",
};
const data = reactive({
  config: {},
  // formCfgOtherInfo: {},
  record: props.modelValue,
  loading: false,
  closeAttch: null,
  // loadingOther: false,
});

const frmCtl = ref(null);

const axios = inject("axios");
function setLookupUrlTrayekID() {
  frmCtl.value.setFieldAttr(
    "TrayekID",
    "lookupUrl",
    `/bagong/trayek/find?Dimension.Key=${props.dimension[2].Key}&Dimension.Value=${props.dimension[2].Value}`
  );
}
function setHideTrayekID(hide) {
  frmCtl.value.setFieldAttr("TrayekID", "hide", hide);
  data.record.DetailUnit.TrayekID =
    data.record.DetailUnit.TrayekID.length > 0
      ? data.record.DetailUnit.TrayekID
      : "";
}
function formOpen() {
  util.nextTickN(2, () => {
    setHideTrayekID(data.record.DetailUnit.Purpose !== "Trayek");
    setLookupUrlTrayekID();
  });
}

function getConfig(url) {
  data.loading = true;
  if (url == "" || props.groupID == null) return;
  loadFormConfig(axios, url).then(
    (r) => {
      // util.nextTickN(2, () => {
      //   const valRef = formDetailAsset.value;
      //   valRef.setFieldAttr("TrayekID", "hide", true);

      //   // beautify form
      //   // const elements = document.querySelectorAll(
      //   //   ".form_detail_asset_bagong .gridCol4"
      //   // );
      //   // if (elements.length > 0) {
      //   //   elements[0].classList.remove("gridCol4");
      //   //   elements[0].classList.add("gridCol3");
      //   //   elements[0].classList.add("trk-hide");
      //   // }
      // });

      // initialize form config
      data.config = r;
      data.loading = false;
      formOpen();
    },
    (e) => {
      util.showError(e);
    }
  );

  // loadFormConfig(axios, "/bagong/asset/detail/assetotherinfo/formconfig").then(
  //   (r) => {
  //     data.formCfgOtherInfo = r;
  //     data.loadingOther = false;
  //   },
  //   (e) => util.showError(e)
  // );
}

const mapUrl = {
  ELC: "electronic",
  PRT: "property",
  UNT: "unit",
};

watch(
  () => props.groupID,
  (nv) => {
    let url = "/bagong/asset/detail/" + mapUrl[nv] + "/formconfig";
    getConfig(url);
  },
  { deep: true }
);

watch(
  () => data.record,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
function fieldChange(name, v1, v2, old) {
  if (name === "Purpose") {
    setHideTrayekID(v1 !== "Trayek");
  }
}
// watch(
//   () => data.record,
//   (nv) => {
//     emit("update:modelValue", nv);
//     const valRef = formDetailAsset.value;
//     if (nv.DetailUnit.Purpose == "Trayek") {
//       valRef.setFieldAttr("TrayekID", "hide", false);
//     } else {
//       valRef.setFieldAttr("TrayekID", "hide", true);
//     }
//   },
//   { deep: true }
// );

onMounted(() => {
  getConfig("/bagong/asset/detail/" + mapUrl[props.groupID] + "/formconfig");
});
</script>
<style>
.s-form-assetdetail.DetailUnit
  .section_group:nth-child(1)
  > .section:nth-child(1)
  > .section_title {
  width: 200%;
}
.s-form-assetdetail.DetailUnit
  .section_group:nth-child(1)
  > .section:nth-child(2)
  > .section_title {
  color: transparent;
}
</style>
