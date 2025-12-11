<template>
  <div>
    <s-form
      ref="form"
      :config="data.config"
      v-model="data.value"
      keep-label
      @field-change="onFieldChange"
    >
      <template #buttons>&nbsp;</template>
    </s-form>
  </div>
</template>

<script setup>
import { reactive, onMounted, inject, ref } from "vue";
import { SForm, loadFormConfig, util } from "suimjs";

const props = defineProps({
  id: { type: String, default: "" },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const data = reactive({
  config: {},
  value: {
    CustomerJournalTypeID: props.id,
    DividerType: "Manual",
    Divider: 0,
    StandbyRate: 0,
    WorkingHour: 0,
  },
});

const form = ref(null);
const axios = inject("axios");

onMounted(() => {
  loadFormConfig(
    axios,
    "/bagong/customerjournaltypeconfiguration/formconfig"
  ).then(
    (r) => {
      data.config = r;
    },
    (e) => util.showError(e)
  );

  axios
    .post(
      "/bagong/customerjournaltypeconfiguration/find?CustomerJournalTypeID=" +
        props.id
    )
    .then(
      (r) => {
        if (r.data && r.data.length > 0) {
          if (r.data[0].DividerType == "Auto") {
            form.value.setFieldAttr("Divider", "hide", true);
          } else {
            form.value.setFieldAttr("Divider", "hide", false);
          }
          data.value = r.data[0];
        }
      },
      (e) => {
        data.loading = false;
      }
    );
});

function onFieldChange(name, v1, v2, old) {
  if (name == "DividerType") {
    if (v1 == "Auto") {
      form.value.setFieldAttr("Divider", "hide", true);
      data.value.Divider = 0;
    } else {
      form.value.setFieldAttr("Divider", "hide", false);
    }
  }
}

function getDataValue() {
  return data.value;
}

defineExpose({
  getDataValue,
});
</script>
