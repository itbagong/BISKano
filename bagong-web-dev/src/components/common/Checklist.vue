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
			:grid-fields="['Key', 'PIC']"
			@grid-row-add="newRecord"
			@grid-refreshed="refresh"
			@grid-row-delete="deleteRecord" 
			@alterGridConfig="onAlterGridConfig"
		>
			<template #grid_Key="props">
				<s-input
					v-model="props.item.Key"
					:readOnly="readOnly || (isKeySecured && props.item.isTemplate)"
				/>
			</template>
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
					:kind="`${attchKind}`"
					:ref-id="attchRefId" 
					:tags="[`${attchTagPrefix}_CHECKLIST_${attchRefId}_${item.Key}`]"
					:read-only="readOnly"
					@preOpen="emit('preOpenAttch', readOnly)"
					@close="emit('closeAttch', readOnly)"
				/>
			</template>
			<template #grid_item_button_delete="{item}">
				<template v-if="isKeySecured && item.isTemplate">&nbsp;</template>
			</template>
			<template #grid_paging>&nbsp;</template>
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
	hideFields: {type: Array, default: () => []},
	isKeySecured: {type: Boolean, default: false},
});

const emit = defineEmits({
	"update:modelValue": null,
	recalc: null,
	preOpenAttch: null,
  closeAttch: null,
  reOpen:null
});

const data = reactive({
	records: props.modelValue !== null ? props.modelValue : [],
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

function changeChecklistID(id) {
	if (id == "") { 
		refresh();
		return;
	}

	axios.post("/tenant/checklisttemplate/get", [id]).then(
		(r) => {
			const newRecords = [];
			if (r.data && r.data.Checklists && r.data.Checklists.length > 0) {
				r.data.Checklists.forEach((el) => {
					const existings = data.records.filter((rcd) => rcd.Key == el.Key);
					if (existings.length > 0) {
						existings[0].suimRecordChange = false;
						existings[0].isTemplate = true
						newRecords.push(existings[0]);
					} else {
						el.Expected = null;
						el.Actual = null;
						el.isTemplate = true
						el.suimRecordChange = false;
						newRecords.push(el);
					}
				});
			}

			data.records.forEach((el) => {
				const existings = newRecords.filter((rcd) => rcd.Key == el.Key);
				if (existings.length == 0) {
					el.suimRecordChange = false;
					newRecords.push(el);
				}
			});

			data.records = newRecords;
			refresh();
		},
		(e) => {}
	);
}
const hasAttch = computed({
  get() {
    return props.attchKind != "" && props.attchRefId != "";
  },
});
watch(
	() => props.checklistId,
	(nv) => { 
		changeChecklistID(nv);
	},
	{deep: true}
);
function onAlterGridConfig(cfg) {
	if (props.hideFields.length > 0) {
		cfg.fields = cfg.fields.filter((o) => !props.hideFields.includes(o.field))
	}
	setTimeout(()=>{
	changeChecklistID(props.checklistId);
	},500)
}
onMounted(() => {
	
});
</script>
