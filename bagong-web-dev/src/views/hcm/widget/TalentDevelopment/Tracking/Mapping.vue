<template>
  <s-form
    v-model="data.record"
    :config="data.formCfg"
    :mode="'view'"
    keep-label
    hide-submit
    hide-cancel
    ref="formRef"
  >
  </s-form>
</template>
<script setup>
import { reactive, onMounted, inject, ref } from "vue";
import { SForm, loadFormConfig, util } from "suimjs";

const axios = inject("axios");
const formRef = ref(null);
const props = defineProps({
  employeeID: { type: String, default: "" },
});
const data = reactive({
  employeeID: props.employeeID,
  formCfg: {},
  record: {},
});

const getsData = () => {
  axios.post("/bagong/employee/get", [data.employeeID]).then(
    (r) => {
      data.record = {
        ...r.data.Detail,
        Name: r.data.Name,
        NIK: r.data.Detail.IdentityCardNo,
        Group: r.data.Detail.IdentityCardNoGroup,
        Site: r.data.Dimension.find((o) => o.Key === "Site")?.Value,
      };
      loadFormConfig(axios, "/hcm/talentdevelopment/mapping/formconfig").then(
        (r) => {
          data.formCfg = r;
          util.nextTickN(2, () => {});
        },
        (e) => util.showError(e)
      );
    },
    (e) => {
      util.showError(e);
    }
  );
};
onMounted(() => {
  getsData();
});
</script>
