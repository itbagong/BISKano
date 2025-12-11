<template>
    <div class="flex flex-col gap-2">
        <data-list
				ref="ContractChecklistGrid" 
                no-gap
				form-keep-label
				grid-hide-sort
                grid-hide-select 
                grid-hide-new
                grid-hide-search 
                grid-hide-refresh 
                grid-hide-detail
                grid-hide-delete
                gridAutoCommitLine
                grid-no-confirm-delete
                grid-config="/sdp/contract-checklist/checked/gridconfig"
				grid-mode="grid"
                init-app-mode="grid" 
                :grid-fields="[
                    'UnitType',
                    'SunID',
                    'AssetID',
                    'UnitStatus',
                    'DocumentStatus',
                    'DeliveryStatus',
                    'CommisioningDate',
                    'CommisioningResult',
                    'CommisioningStatus'
                ]"
                @grid-row-field-changed="onGridRowFieldChanged"
                @alter-grid-config="alterGridConfig"
			>
            <template #grid_CommisioningDate="{item}">
                <s-input
                    kind="date"
                    v-model="item.CommisioningDate"
                    class="min-w-[100px]"
                >
                </s-input>
            </template>
            <template #grid_AssetID="{item}">
                <span>{{item.AssetName}}</span>
            </template>
            <template #grid_CommisioningResult="{item}">
                <s-input
                    v-model="item.CommisioningResult"
                    class="min-w-[100px]"
                    @change="(old, v1, v2) => handleChangeCommisioningResult(v1, item.CommisioningResult)"
                >
                </s-input>
            </template>
            <template #grid_DeliveryStatus="{item}">
                <s-input
                    use-list
                    v-model="item.DeliveryStatus"
                    class="min-w-[100px]"
                    :items="['On Progress', 'Delivered']"
                >
                </s-input>
            </template>
            <template #grid_CommisioningStatus="{item}">
                <s-input
                    use-list
                    v-model="item.CommisioningStatus"
                    class="min-w-[100px]"
                    :lookup-url="'tenant/masterdata/find?MasterDataTypeID=CMS'"
                    lookup-key="_id"
                    :lookup-labels="['Name']"
                >
                </s-input>
            </template>
            <template #grid_item_buttons_1="{item}">
                <s-button class="bg-blue-400 m-1" v-if="item.CommisioningStatus == 'CMS001'" label="BAST" @click="bastForm(item)"></s-button>
            </template>
		</data-list>
        <!-- <div v-else-if="data.checklistMode == 'bast'">
            {{data.bastRecord}}
           
        </div> -->
    </div>
</template>

<script setup>
import { onMounted, inject } from "vue";
import { reactive, ref } from "vue";
import { DataList, SButton, SInput, util } from "suimjs";


const axios = inject("axios");

const ContractChecklistGrid = ref(null);


const props = defineProps({
    isEdited: {type: Boolean, default: () => false},
    modelValue: { type: Array, default: () => [] },
    salesOrderRefNo: {type: String, default: () => ""},
    salesOrderId: {type: String, default: () => ""},
    openSection: {type: String, default: () => ""},
});

const data = reactive({
  records: props.modelValue,
  salesOrderRefNo: props.salesOrderRefNo,
  salesOrderId: props.salesOrderId,
  isEdited: props.isEdited,
  openSection: props.openSection,
  bastRecord: {},
  checklistMode: 'checklist',
  collection: 
		{
		"section1": [
			{
				key: "O - Ring",
				value: false
			},
			{
				key: "FSI",
				value: false
			},
			{
				key: "AutoLUB",
				value: false
			},
			{
				key: "Blade",
				value: false
			},
			{
				key: "Ripper",
				value: false
			},
			{
				key: "Pontoon",
				value: false
			},
		],
		"section2": [
			{
				key: "Bucket",
				value: false
			},
			{
				key: "AC",
				value: false
			},
			{
				key: "Radio",
				value: false
			},
			{
				key: "Kaca Spion",
				value: false
			},
			{
				key: "Kaca Spion Kiri",
				value: false
			},
			{
				key: "Suction Hose",
				value: false
			},
		],
		"section3": [
			{
				key: "Kaca Spion Dalam",
				value: false
			},
			{
				key: "Lampu Sein Kiri",
				value: false
			},
			{
				key: "Lampu Sein kanan",
				value: false
			},
			{
				key: "Kunci Roda",
				value: false
			},
			{
				key: "Part Manual / Literatur",
				value: false
			},
			{
				key: "Winch",
				value: false
			},
		],
		"section4": [
			{
				key: "Ban Cadangan",
				value: false
			},
			{
				key: "Baterai",
				value: false
			},
			{
				key: "Wiper",
				value: false
			},
			{
				key: "Kunci Kontak",
				value: false
			},
			{
				key: "Fire Extuingiser",
				value: false
			},
			{
				key: "Lain-lain",
				value: false
			},
		]
	}
});

const emit = defineEmits({
  "update:modelValue": null,
  "update:openSection": null,
  "update:itemSelected": null,
  recalc: null,
});


function onGridRowFieldChanged(name, v1, v2, old, record) {
    
//   updateItems();
}

function handleChangeCommisioningResult(value, item) {
}

function updateItems() {
  const record = data.records
  emit("update:modelValue", record);
  emit("recalc");
}

async function showBySUNID(sunID) {
    let data = []
    let documentStatus = ""
    await axios.post('/sdp/documentunitchecklist/find?SUNID='+sunID).then((res) => {
        const response = res.data[0]
        
        if (
            response.StatusSRUT == "" &&
            response.StatusRekomPeruntukan == "" &&
            response.StatusSamsat == "" &&
            response.StatusUjiKIR == "" &&
            response.StatusFinal == ""
        ) {
            documentStatus = "Need Action"
        }else if (response.StatusFinal != "") {
            documentStatus = response.StatusFinal
        }else if (response.StatusSRUT != "") {
            documentStatus = response.StatusSRUT
        }else if (response.StatusRekomPeruntukan != "") {
            documentStatus = response.StatusRekomPeruntukan
        }else if (response.StatusSamsat != "") {
            documentStatus = response.StatusSamsat
        } else if (response.StatusUjiKIR != "") {
            documentStatus = response.StatusUjiKIR
        } else if (response.StatusRoutePermit != "") {
            documentStatus = response.StatusRoutePermit
        }

        data = response
    })

    return {
        data: data,
        documentStatus: documentStatus
    }
}

function alterGridConfig(item) {
    
}

async function getWorkOrder() {
    const res = data.salesOrderId != "" ? 
        await axios.post('/mfg/work/request/gets', 
        {
            "Where": {
                "Field": "SourceID",
                "Op": "$eq",
                "Value": data.salesOrderId
            }
         }).then((res) => {
            return res.data.data
        }) : []
    if (props.isEdited == true) {
        let items = []
        // console.log("==>", ContractChecklistGrid.value.getGridField())
        ContractChecklistGrid.value.setGridRecords([])
        if (props.modelValue.length > 0) {
             props.modelValue.forEach((v, i) => {
                let isBast = true
                if (!Object.keys(v.Bast).length) {
                    isBast = false
                }
                let collection = {}
                if (v.Bast == null) {
                    collection = data.collection
                }else {
                    collection = v.Bast
                }

                items.push({  
                    UnitType: v.UnitType,
                    SunID: v.SunID,
                    AssetID: v.AssetID,
                    AssetName: v.AssetName,
                    UnitStatus: v.UnitStatus,
                    DocumentStatus: v.DocumentStatus,
                    DeliveryStatus: v.DeliveryStatus,
                    CommisioningDate: v.CommisioningDate,
                    CommisioningResult: v.CommisioningResult,
                    CommisioningStatus: v.CommisioningStatus,
                    isBast: isBast,
                    Bast: {
                        collection: collection
                    },
                })
            })
        }

        data.records = items
        ContractChecklistGrid.value.setGridRecords(items)
        updateItems()
        return
    }else {
        data.records = []
        ContractChecklistGrid.value.setGridRecords([])
        if (props.modelValue.length > 0) {
            props.modelValue.forEach((v, i) => {
                let documentStatus = ""
                if (v.SunID != "") {
                    axios.post('/sdp/documentunitchecklist/find?SUNID='+v.SunID).then((res) => {
                        const response = res.data[0]
                        if (
                            response.StatusSRUT == "" &&
                            response.StatusRekomPeruntukan == "" &&
                            response.StatusSamsat == "" &&
                            response.StatusUjiKIR == "" &&
                            response.StatusFinal == ""
                        ) {
                            documentStatus = "Need Action"
                        }else if (response.StatusFinal != "") {
                            documentStatus = response.StatusFinal
                        }else if (response.StatusSRUT != "") {
                            documentStatus = response.StatusSRUT
                        }else if (response.StatusRekomPeruntukan != "") {
                            documentStatus = response.StatusRekomPeruntukan
                        }else if (response.StatusSamsat != "") {
                            documentStatus = response.StatusSamsat
                        } else if (response.StatusUjiKIR != "") {
                            documentStatus = response.StatusUjiKIR
                        } else if (response.StatusRoutePermit != "") {
                            documentStatus = response.StatusRoutePermit
                        }

                        data.records.push({
                            _id: v._id,
                            UnitType: v.UnitType,
                            SunID: v.SunID,
                            AssetID: v.AssetID,
                            AssetName: v.AssetName,
                            UnitStatus: v.UnitStatus,
                            DocumentStatus: documentStatus,
                            DeliveryStatus: v.DeliveryStatus,
                            CommisioningDate: v.CommisioningDate,
                            CommisioningResult: v.CommisioningResult,
                            CommisioningStatus: v.CommisioningStatus,
                            isBast: v.isBast,
                            Bast: {
                                collection: data.collection
                            },
                        })
                    })
                    
                }else {
                    data.records.push({
                        _id: v._id,
                        UnitType: v.UnitType,
                        SunID: v.SunID,
                        AssetID: v.AssetID,
                        AssetName: v.AssetName,
                        UnitStatus: v.UnitStatus,
                        DocumentStatus: v.DocumentStatus,
                        DeliveryStatus: v.DeliveryStatus,
                        CommisioningDate: v.CommisioningDate,
                        CommisioningResult: v.CommisioningResult,
                        CommisioningStatus: v.CommisioningStatus,
                        isBast: v.isBast,
                        Bast: {
                            collection: data.collection
                        },
                    })
                }
            });
        }else {
            if (res.length > 0) {
                res.forEach((v, i) => {
                    console.log("v =>", v._id)
                    axios.post("/mfg/workorderplan/find?WorkRequestID="+v._id).then((res) => {
                        var woData = res.data[0]
                        let documentChecklist = showBySUNID(woData.SunID)
                        console.log("woData.Asset =>", woData.Asset)
                        let assetTenant = axios.post("/tenant/asset/find?_id="+woData.Asset).then((resAsset) => {
                            return resAsset.data[0]
                        })
                        Promise.all([documentChecklist, assetTenant]).then(([documentChecklistRecord, assetData]) => {
                            console.log("documentChecklist =>", documentChecklistRecord)
                            // axios.post("/tenant/asset/find?_id="+woData.Asset).then((resAsset) => {
                            //     var assetData = resAsset.data[0]

                                data.records.push({
                                    _id: v._id,
                                    UnitType: v.WorkRequestType,
                                    SunID: woData.SunID,
                                    AssetID: woData.Asset,
                                    AssetName: woData.Asset + " | " + assetData.Name,
                                    UnitStatus: woData.StatusOverall,
                                    DocumentStatus: documentChecklistRecord.documentStatus,
                                    DeliveryStatus: "",
                                    CommisioningDate: "0001-01-01T00:00:00Z",
                                    CommisioningResult: "",
                                    CommisioningStatus: "",
                                    isBast: false,
                                    Bast: {
                                        collection: data.collection
                                    },
                                })
                            // })
                        })
                        
                    })
                })
            }
        }
        
        // console.log("data records =>", data.records)
        ContractChecklistGrid.value.setGridRecords(data.records)
        updateItems()
   }
    // data.recordChecklist = dataModel
    // data.record = data

}

function bastForm(item) {
    data.checklistMode = "bast"
    data.openSection = "bast-form"
    // console.log("after props.openSection =>", data.openSection)
    item.isBast = true

    emit("update:itemSelected", bastForm, item, props.isEdited)
    emit("update:openSection", data.openSection)
    emit("recalc");
    // console.log("Status =>", item)

}

function setGridRecords(items) {
  ContractChecklistGrid.value.setGridRecords(items);
}

onMounted(() => {
    if (props.isEdited == true) {
        window.setTimeout(() => {
             getWorkOrder()
        }, 2000)
    }
    getWorkOrder()
    
    // console.log("props.openSection =>",  data.openSection)
});

defineExpose({
    setGridRecords,
})
</script>