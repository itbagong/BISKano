<template>
  <s-modal title="Select Company" display hideTitle hideButtons hideClose>
    <template #title>
        <div class="p-4 pb-0">
        <h1 class="text-primary border-b-[1px]">Select Company</h1>
        </div>
    </template>
    <s-card hide-title  class="min-w-[400px]"> 
        <div class="py-3 mb-3 ">
            <ul class="flex flex-col">
                <li
                    v-for="(company,idx) in items"
                    :key="idx"
                    @click="data.selected = data.selected == company._id ? '' : company._id"
                    class=" border-b border-slate-300   cursor-pointer hover:bg-[#f3e2e3] flex justify-between items-center p-1 py-2
                    " 
                    :class="[data.selected == company._id?'bg-[#f3e2e3]':'']"
                >
                    <div  >{{ company.Name }}  </div> 
                    <mdicon v-if=" data.selected == company._id" name="check-circle"  class="text-primary" size="14"/> 
                </li>
            </ul>
        </div>
        <template #footer>
            <div class="mt-5">
              <s-button class="bg-primary text-white font-bold w-full" label="Submit" @click="changeCompany"  :disabled="data.selected == ''"  ></s-button>
            </div>
        </template>
    </s-card>
  </s-modal>
</template>
<script setup>
import { onMounted, reactive, inject } from "vue";
import { SCard, util, SModal, SButton } from "suimjs";
import { authStore } from "@/stores/auth";

const auth = authStore();
const axios = inject("axios");

const props = defineProps({
  items: { type: Array, default: [] },
})
const emit = defineEmits({
  close: null,
  submit:null,
});
const data = reactive({
  companies: [],
  selected: auth.appData.CompanyID,
});
function changeCompany(){
  emit("submit", data.selected)
  emit("close")
}
</script>