<template>
	<div>
		<data-list
			ref="listControl"
			hide-title
			no-gap 
			:grid-editor="!readOnly"
			:grid-hide-delete="readOnly"
			:grid-hide-control="readOnly"
			grid-hide-select
			grid-hide-search
			grid-hide-detail
			grid-hide-sort
			grid-hide-refresh
			grid-no-confirm-delete
			gridAutoCommitLine
			init-app-mode="grid"
			grid-mode="grid"
			form-keep-label
			new-record-type="grid"
			grid-config="/tenant/checklistitem/gridconfig"
			:grid-fields="['PIC']"
			@grid-row-add="newRecord"
			@grid-refreshed="refresh"
			@grid-row-delete="deleteRecord" 
			@alterGridConfig="onAlterGridConfig"
		>
			<template #grid_PIC="props">
				<s-input
					v-model="props.item.PIC"
					useList
					lookup-key="_id"
					:lookup-labels="['DisplayName']"
					:lookupSearchs="['_id', 'DisplayName']"
					lookupUrl="/iam/user/find-by"
				/>
			</template>
			<template #grid_item_buttons_1="{item}"> 
				<action-attachment
					v-if="hasAttch && item.Key"  
					:kind="`${attchKind}`"
					:ref-id="attchRefId" 
					:tags="[`${attchTagPrefix}_CHECKLIST_${attchRefId}_${item.Key}`]"
					:read-only="readOnly"
					@preOpen="emit('preOpenAttch', readOnly)"
					@close="emit('closeAttch', readOnly)"
				/>
			</template>
		</data-list>
	</div>
</template>

<script setup>
import {reactive, inject, ref, onMounted, watch, computed} from "vue";
import {DataList, SInput, util} from "suimjs";
import ActionAttachment from "@/components/common/ActionAttachment.vue";

const props = defineProps({
	modelValue: {type: Array, default: () => []},
	readOnly: {type: Boolean, default: false},
	checklistId: {type: String, default: ""},
	attchKind: {
    type: String,
    default: "",
  },
  attchRefId: {
    type: String,
    default: "",
  },
  attchTagPrefix: {
    type: String,
    default: "",
  },
});

const emit = defineEmits({
	"update:modelValue": null,
	recalc: null,
	preOpenAttch: null,
  closeAttch: null,
});

const data = reactive({
	records: props.modelValue,
});

function newRecord() {
	const record = {};
	record.suimRecordChange = false;
	record.Expected = null;
	record.Actual = null;
	data.records.push(record);
	refresh();
}

function deleteRecord(rcd, index) {
	const newRecords = data.records.filter((dt, idx) => {
		return idx != index;
	});
	data.records = newRecords;
	refresh();
}

function refresh() {
	listControl.value.setGridRecords(data.records);
	emit("update:modelValue", data.records);
}

const axios = inject("axios");
const listControl = ref(null);

 
function onAlterGridConfig(cfg) {
	setTimeout(()=>{
		refresh()
	},500)
}
const hasAttch = computed({
  get() {
    return props.attchKind != "" && props.attchRefId != "";
  },
});
defineExpose({
	refresh
})
</script>
