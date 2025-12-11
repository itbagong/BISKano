<template>
    <div class="w-full flex mb-4 gap-2">
        <div class="flex justify-end grow gap-2">
            <s-button class="btn_secondary hover:bg-white hover:text-primary" icon="keyboard-backspace"
                @click="cancelForm"></s-button>
            <h1 class="grow">Create New Submision of Petty Cash</h1>
            <s-button class="bg-transparent hover:bg-blue-500 hover:text-black" label="Preview" icon="eye-outline"
                @click="saveForm"></s-button>
            <s-button class="transparent hover:bg-green-500 hover:text-black" label="Request Approve"
                icon="archive-check-outline" @click="saveForm"></s-button>
            <s-button class="btn_primary hover:bg-white hover:text-primary" label="Save" icon="content-save"
                @click="saveForm"></s-button>
        </div>
    </div>
    <div class="w-full">
        <s-form ref="formGeneral" class="submission-detail" :config="data.formConfig" v-model="data.formRecord"
            :tabs="['General', 'Lines', 'SubmissionHistory', 'Attachment']" hideButtons hideSubmit hideCancel
            @submitForm="saveForm">
            <template #input_Dimension="{ item }">
                <DimensionEditor v-model="item.Dimension"></DimensionEditor>
            </template>
            <template #input_Recurrence="{ item }">
                <RecurrenceEditor v-model="item.Recurence"></RecurrenceEditor>
            </template>
            <template #tab_Lines>
                <s-grid ref="gridLines" :config="data.gridLinesConfig" :modelValue="data.gridLinesRecords"
                    :editor="data.editor" hideDetail hideControl hideSaveButton noConfirmDelete @deleteData="deleteData"
                    @rowFieldChanged="rowFieldChanged">
                </s-grid>
                <s-button class="text-red-500 font-bold" icon="plus-box-outline" label="Add more items"
                    @click="addNewLines"></s-button>
            </template>
            <template #tab_SubmissionHistory>
                Submission History yes
            </template>
            <template #tab_Attachment>
                Attachment yes
            </template>
        </s-form>
    </div>
</template>
<script setup>
import { reactive, ref, onMounted, inject, nextTick } from "vue";
import { DataList, SGrid, SInput, SButton, SModal, SForm, util, createFormConfig } from "suimjs";
import DimensionEditor from "@/components/common/DimensionEditor.vue";
import RecurrenceEditor from "@/components/common/RecurrenceEditor.vue";
import { authStore } from '@/stores/auth';
import moment from "moment"

const auth = authStore()
const axios = inject("axios");

const formGeneral = ref(null);
const gridLines = ref(null);

const data = reactive({
    formConfig: {},
    formRecord: {
        SubmissionDate: "",
    },
    gridLinesConfig: {
        fields: [],
        setting: {}
    },
    gridLinesRecords: [],
    gridLinesRecord: {},
    editor: true,
    noConfirmDelete: true,
})

const props = defineProps({
    dataParameter: {},
});

const emit = defineEmits({
    cancelForm: null,
})

function cancelForm() {
    emit("cancelForm")
}

onMounted(() => {
    generateFormConfig()
    if (Object.keys(props.dataParameter).length !== 0) {
        data.formRecord = props.dataParameter
    }
    generateConfigLines()
    fetchDataLine()
})

function generateFormConfig() {
    const cfg = createFormConfig("", true);
    cfg.addSection("General Info", true).addRowAuto(2, {
        field: "SubmissionNo",
        kind: "text",
        label: "Submission No",
        required: true,
    }, {
        field: "SubmissionTitle",
        kind: "text",
        label: "Submission Title",
        required: true,
    }, {
        field: "JurnalType",
        label: "Jurnal Type",
        required: true,
        useList: true,
        allowAdd: false,
        lookupKey: "_id",
        lookupLabels: ["_id", "Name"],
        lookupSearch: ["_id", "Name"],
        lookupUrl: "/fico/vendorjournaltype/find",
    }, {
        field: "InclusiveTax",
        kind: "checkbox",
        label: "Inclusive Tax",
        required: false,
    }, {
        field: "SubmissionDate",
        kind: "date",
        label: "Submission Date",
        required: true,
        readOnly: false,
    },
    );
    data.formConfig = cfg.generateConfig();

    cfg.addSection("Cash Bank", true).addRowAuto(2, {
        field: "FromCashBankID",
        kind: "text",
        label: "From Cash Bank",
        required: true,
        readOnly: false,
        useList: true,
        lookupKey: "_id",
        lookupLabels: ["Name"],
        lookupSearchs: ["_id", "Name"],
        lookupUrl: "/tenant/cashbank/find"
    }, {
        field: "ToCashBankID",
        label: "To Cash Bank",
        show: true,
        kind: "text",
        readOnly: false,
        useList: true,
        lookupKey: "_id",
        lookupLabels: ["Name"],
        lookupSearchs: ["_id", "Name"],
        lookupUrl: "/tenant/cashbank/find"
    })

    cfg.addSection("Dimension", true).addRowAuto(2, {
        field: "Dimension",
        kind: "input",
        label: "Name",
        required: true,
    })
    data.formConfig = cfg.generateConfig();

    // cfg.addSection("Recurrence", true).addRowAuto(2, {
    //     field: "Recurrence",
    //     kind: "input",
    //     label: "Recurrence",
    //     required: true,
    // })
    data.formConfig = cfg.generateConfig();
}

function saveForm() {
    formGeneral.value.setLoading(true)
    let payload = data.formRecord
    payload.CompanyID = auth.appData.CompanyID
    axios.post("/bagong/pettycash_submission/save", payload)
        .then(async r => {
            util.showInfo('Data has been saved');
            gridLines.value.refreshData();
        }, e => {
            util.showError(e);
        }).finally(() => {
            if (formGeneral.value) {
                formGeneral.value.setLoading(false);
            }
        });
}

function generateConfigLines() {
    let columns = [
        {
            field: "Asset",
            label: "Asset",
            show: true,
            kind: "text",
            readOnly: false,
            useList: true,
            lookupLabels: ["Name"],
            lookupSearch: ["_id", "Name"],
            lookupUrl: "/tenant/asset/find"
        },
        {
            field: "Description", label: "Description", show: "show", kind: "text", readOnly: false
        }, {
            field: "Quantity", label: "Quantity", show: true, kind: "number", readOnly: false
        }, {
            field: "UoM", label: "UoM", show: true, kind: "text", readOnly: false
        }, {
            field: "UnitPrice", label: "UnitPrice", show: true, kind: "number", readOnly: false
        }, {
            field: "Total", label: "Total", show: true, kind: "number", readOnly: true
        }, {
            field: "ApprovedAmount", label: "Approved Amount", show: true, kind: "number", readOnly: false
        }, {
            field: "Urgent", label: "Urgent", show: true, kind: "text", readOnly: false
        }, {
            field: "Ledger", label: "Ledger", show: true, kind: "text", readOnly: false
        },
        {
            field: "CashBank",
            label: "CashBank",
            show: true,
            kind: "text",
            readOnly: false,
            useList: true,
            lookupLabels: ["Name"],
            lookupSearchs: ["_id", "Name"],
            lookupUrl: "/tenant/cashbank/find"
        },
        {
            field: "Critical", label: "Critical", show: true, kind: "text", readOnly: false
        }]

    let fields = []
    columns.map((x, i) => {
        fields.push({
            "field": x.field,
            "label": x.label,
            "halign": "start",
            "valign": "start",
            "labelField": "",
            "readType": x.show ? "show" : "hide",
            "input": {
                "field": x.field,
                "useList": x.useList,
                "lookupKey": "_id",
                "lookupLabels": x.lookupLabels,
                "lookupSearchs": x.lookupSearch,
                "lookupUrl": x.lookupUrl,
                "readOnly": x.readOnly,
                "kind": x.kind
            },
            "slot": "item_" + x.field,
            "kind": x.kind,
        })
        data.gridLinesRecord[x.field] = ""
    })

    data.gridLinesConfig.fields = fields
}

function fetchDataLine() {
    let dataLines = props.dataParameter.Detail
    data.gridLinesRecords = dataLines
}

function addNewLines() {
    const record = {
        "Asset": "",
        "Description": "",
        "Quantity": 1,
        "UoM": "",
        "UnitPrice": 1,
        "Total": 1,
        "Urgent": "",
        "Ledger": "",
        "CashBank": "",
        "Critical": ""
    }
    // data.gridLinesRecords.push(record)
    gridLines.value.addData(record)
    gridLines.value.newData();
}

function deleteData(record, index) {
    record.items.splice(index, 1)
    fetchDataLine()
}

function rowFieldChanged(name, v1, v2, old, record) {
    if (name == "Quantity") {
        record.Total = v1 * record.UnitPrice
    }
    if (name == "UnitPrice") {
        record.Total = v1 * record.Quantity
    }
    gridLines.value.setRecord(
        record,
        gridLines.value.getCurrentIndex()
    );
}
</script>

<style></style>