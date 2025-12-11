<template>
    <div class="w-full flex mb-4 gap-2">
        <div class="flex justify-end grow gap-2">
            <s-button class="btn_secondary hover:bg-white hover:text-primary" icon="keyboard-backspace"
                @click="cancelForm"></s-button>
            <h1 class="grow">Create New Submision of Employee Expense</h1>
            <s-button class="bg-transparent hover:bg-blue-500 hover:text-black" label="Preview" icon="eye-outline"
                @click="saveForm"></s-button>
            <s-button class="transparent hover:bg-green-500 hover:text-black" label="Request Approve"
                icon="archive-check-outline" @click="saveForm"></s-button>
            <s-button class="btn_warning hover:bg-white hover:text-primary" label="Save" icon="content-save"
                @click="saveSubmission"></s-button>
        </div>
    </div>
    <div class="w-full">
        <s-form class="submission-detail" :config="data.formConfig" v-model="data.formRecord"
            :tabs="['General', 'Lines', 'Attachment']" hideButtons hideSubmit hideCancel @submitForm="saveForm"
            :loading="data.formLoading">
            <template #input_Dimension="{ item }">
                <DimensionEditor v-model="item.Dimension"></DimensionEditor>
            </template>
            <template #tab_Lines>
                <s-grid ref="gridLines" :config="data.gridLinesConfig" :modelValue="data.gridLinesRecords"
                    :editor="data.isEditor" hideDetail hideControl hideSaveButton noConfirmDelete @deleteData="deleteData">
                    <!-- <template #item_Asset="item">
                        {{ item }}
                    </template> -->
                    <!-- <template v-for="(col, index) in data.gridLinesConfig.fields" v-slot:[col.slot]="item">
                        <div v-if="!item.item.isNew">{{ item.item[item.header.field] }}</div>
                        <div v-else>
                            <s-input field="{{item.header.field}}" :kind="item.header.kind"
                                v-model="item.item[item.header.field]" hide-label></s-input>
                        </div>
                    </template> -->
                </s-grid>
                <s-button class="text-red-500 font-bold" icon="plus-box-outline" label="Add more items"
                    @click="addNewLines"></s-button>
                <!-- <s-button class="text-red-500 font-bold" icon="plus-box-outline" label="fetch"
                    @click="fetchDataLine"></s-button> -->
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
// import RecurrenceEditor from "@/components/common/RecurrenceEditor.vue";
import { authStore } from '@/stores/auth';
import moment from "moment"

const auth = authStore()
const axios = inject("axios");
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
    formLoading: true,
    isEditor: true,
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
    // console.log(auth.appData)
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
        field: "JournalTypeID",
        label: "Journal Type",
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
        field: "EmployeeID",
        label: "Employee Name",
        required: false,
        useList: true,
        allowAdd: false,
        lookupKey: "_id",
        lookupLabels: ["Name"],
        lookupSearch: ["_id", "Name"],
        lookupUrl: "/tenant/employee/find",

    }, {
        field: "SubmissionDate",
        kind: "date",
        label: "Submission Date",
        required: true,
        readOnly: false,
    },
    );
    data.formConfig = cfg.generateConfig();

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

function saveSubmission() {
    let payload = data.formRecord
    payload.CompanyID = auth.appData.CompanyID
    payload.Status = "Waiting Approved"
    // console.log("payload", payload)
    axios.post("/bagong/employeeexpense_submission/save", payload).then(async r => {
        util.showInfo('Data has been saved');
        fetchDataLine()
        // console.log(props.dataParameter.Detail)
        gridLines.value.refreshData();
    }, e => {
        util.showError(e);
    });
}

function generateConfigLines() {
    let columns = [
        // {
        //     field: "Asset",
        //     label: "Asset",
        //     show: true,
        //     kind: "text",
        //     readOnly: false,
        //     useList: true,
        //     lookupLabels: ["_id", "Name"],
        //     lookupSearch: ["_id", "Name"],
        //     lookupUrl: "/tenant/asset/find"
        // }, 
        {
            field: "Description", label: "Description", show: "show", kind: "text", readOnly: false
        }, {
            field: "Quantity", label: "Quantity", show: true, kind: "number", readOnly: false
        }, {
            field: "UoM", label: "UoM", show: true, kind: "text", readOnly: false
        }, {
            field: "UnitPrice", label: "Unit Price", show: true, kind: "number", readOnly: false
        }, {
            field: "Total", label: "Total", show: true, kind: "number", readOnly: true
        }, {
            field: "OffsetAccount",
            label: "Offset Account",
            show: true,
            kind: "text",
            readOnly: false,
            useList: true,
            lookupLabels: ["_id", "Name"],
            lookupSearch: ["_id", "Name"],
            lookupUrl: "/tenant/ledgeraccount/find"
        }, {
            field: "CashBank",
            label: "Cash Bank",
            show: true,
            kind: "text",
            readOnly: false,
            useList: true,
            lookupLabels: ["_id", "Name"],
            lookupSearch: ["_id", "Name"],
            lookupUrl: "/tenant/cashbank/find"
        }, {
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
                "lookupSearch": x.lookupSearch,
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
        "OffsetAccount": "",
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
</script>

<style></style>