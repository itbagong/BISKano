<template>
  <s-modal title="Rejection" display hideTitle hideButtons hideClose>
    <template #title>
      <div class="p-4 pb-0">
        <h1 class="text-primary border-b-[1px] flex justify-between">
          Add Material
          <a href="#" class="delete_close" @click="close()">
            <mdicon
              name="close"
              width="16"
              alt="close"
              class="cursor-pointer hover:text-primary"
            />
          </a>
        </h1>
      </div>
    </template>
    <s-card hide-title class="min-w-[400px]">
      <div class="flex flex-col gap-2">
        <div>
          <s-input-sku-item
            ref="refItemVarian"
            label="Item Varian"
            v-model="data.ItemVarian"
            :record="data.material"
            :required="true"
            :keepErrorSection="true"
            :lookup-url="`/tenant/item/gets-detail`"
          ></s-input-sku-item>
        </div>
        <div v-if="false">
          <s-input
            ref="refItemID"
            label="Item"
            v-model="data.material.ItemID"
            :disabled="false"
            use-list
            :lookup-url="`/tenant/item/find`"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
            class="w-full"
            :required="true"
            :keepErrorSection="true"
            @change="
              (field, v1, v2, old, ctlRef) => {
                onChangeItemMaterial(v1, v2, data.material);
              }
            "
          ></s-input>
        </div>
        <div v-if="false">
          <s-input
            ref="refSKU"
            label="SKU"
            v-model="data.material.SKU"
            :disabled="false"
            use-list
            :lookup-url="`/tenant/itemspec/gets-info?ItemID=${data.material.ItemID}`"
            lookup-key="_id"
            :lookup-labels="['Description']"
            :lookup-searchs="['_id', 'SKU', 'Description']"
            class="w-full"
            :required="true"
            :keepErrorSection="true"
            @change="
              (field, v1, v2, old, ctlRef) => {
                handleChangeSKU(v1, v2, data.material);
              }
            "
          ></s-input>
        </div>
        <div>
          <s-input
            ref="refRequired"
            kind="number"
            v-model="data.material.Required"
            label="Required"
            :disabled="false"
            class="w-full"
            :required="true"
            :keepErrorSection="true"
          ></s-input>
        </div>
      </div>

      <template #footer>
        <div class="mt-5">
          <s-button
            class="bg-primary text-white font-bold w-full flex justify-center"
            label="Add Material"
            @click="changeWork"
          ></s-button>
        </div>
      </template>
    </s-card>
  </s-modal>
</template>
<script setup>
import { onMounted, reactive, inject, ref } from "vue";
import { SCard, util, SModal, SButton, SInput } from "suimjs";
import { authStore } from "@/stores/auth";
import SInputSkuItem from "../../../scm/widget/SInputSkuItem.vue";
const auth = authStore();
const axios = inject("axios");
const refItemVarian = ref(null);
const refItemID = ref(null);
const refSKU = ref(null);
const refRequired = ref(null);
const props = defineProps({
  modelValue: {
    type: Object,
    defaule: {
      ItemID: "",
      SKU: "",
      Required: 0,
    },
  },
});
const emit = defineEmits({
  "update:modelValue": null,
  close: null,
});
const data = reactive({
  material: props.modelValue,
  ItemVarian: "",
  WorkTitle: "",
});

function onChangeItemMaterial(v1, v2, item) {
  item.SKU = "";
  item.Required = 0;
  if (typeof v1 == "string") {
    axios.post("/tenant/item/get", [v1]).then(
      (r) => {
        item.Item = r.data;
      },
      (e) => {
        util.showError(e);
      }
    );
  }
}
function handleChangeSKU(v1, v2, item) {
  axios.post("/tenant/itemspec/gets-detail", [v1]).then((r) => {
    item.ItemSpec = r.data[0];
  });
}
function close() {
  emit("close");
}

function changeWork() {
  if (refItemVarian.value.validate() && refRequired.value.validate()) {
    emit("changeWork", data.material);
  }
}

onMounted(() => {});
</script>
