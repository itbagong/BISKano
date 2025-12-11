<template>
    <div class="mb-2">
        <label class="input_label">{{ label }}</label>
    </div>
    <div class="flex gap-2">
        <s-input
            class="w-full"
            keep-label
            label="Main balance account"
            v-model="data.settingValue.MainBalanceAccount"
            use-list
            lookup-url="/tenant/ledgeraccount/coa/find"
            lookup-key="_id"
            :lookup-labels="['_id', 'Name']"
            :lookup-searchs="['Name']"
            :read-only="readOnly"
        ></s-input>
        <s-input
            class="w-full"
            keep-label
            label="Deposit account"
            v-model="data.settingValue.DepositAccount"
            use-list
            lookup-url="/tenant/ledgeraccount/coa/find"
            lookup-key="_id"
            :lookup-labels="['_id', 'Name']"
            :lookup-searchs="['Name']" 
            :read-only="readOnly"
        ></s-input>
    </div>
</template>
<script setup>
import {reactive, watch,onMounted} from "vue";
import helper from "@/scripts/helper.js";
import {SInput} from "suimjs";

const props = defineProps({
	modelValue: {type: Array, default: () => []}, 
	label: {type: String, default: 'Setting'},
  readOnly: { type: Boolean, default: false },
});

const emit = defineEmits({
	"update:modelValue": null,
});

const data = reactive({
    settingValue:  props.modelValue == null || props.modelValue == undefined
      ? {}
      : props.modelValue,
});

watch(
	() => data.settingValue,
	(nv) => { 
		emit("update:modelValue", nv);
	},
	{deep: true}
);
 
</script>