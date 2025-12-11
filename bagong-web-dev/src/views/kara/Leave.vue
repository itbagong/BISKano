<template>
    <div>
        <data-list class="card" ref="listControl" grid-mode="grid" form-keep-label title="Leave" 
            grid-config="/kara/leave/gridconfig" 
            grid-read="/kara/leave/gets"
            form-config="/kara/leave/formconfig"
            form-read="/kara/leave/get"
            form-insert="/kara/leave/insert"
            form-update="/kara/leave/update"
            grid-hide-delete
            :grid-hide-new="!profile.canCreate"
            :grid-hide-edit="!profile.canUpdate"
            :form-hide-submit="!data.showSave"
            @form-new-data="newRecord"
            @form-field-change="onFormFieldChange"
            @form-loaded="editRecord"
            :form-fields="['Reason']"
            stay-on-form-after-save
            :grid-custom-filter="data.customFilter"
            @gridResetCustomFilter="resetGridHeaderFilter"
        >
            <template #grid_header_search>
                <grid-header-filter
                    ref="gridHeaderFilter"
                    v-model="data.customFilter"
                    hide-filter-text
                    hide-filter-date
                    hide-filter-status
                    @init-new-item="initNewItemFilter"
                    @pre-change="changeFilter"
                    @change="refreshGrid"
                >
                    <template #filter_1="{ item }">
                        <s-input
                            class="min-w-[200px]"
                            label="User ID"
                            use-list
                            multiple
                            v-model="item.UserID"
                            lookup-url="/iam/user/find"
                            lookup-key="DisplayName"
                            :lookup-labels="['DisplayName']"
                        />
                        <s-input
                            class="w-[200px]"
                            keep-label
                            label="Search Title"
                            v-model="item.Keyword"
                        />
                        <s-input
                            label="Leave from"
                            kind="date"
                            v-model="item.LeaveFrom"
                        />
                        <s-input
                            label="Leave to"
                            kind="date"
                            v-model="item.LeaveTo"
                        />
                    </template>
                </grid-header-filter>
            </template>
            <template #grid_item_buttons_2="{ item }">
                <a
                    href="#"
                    v-if="item.Status === 'Draft'"
                    @click="deleteData(item)"
                    class="delete_action"
                >
                    <mdicon
                        name="delete"
                        width="16"
                        alt="delete"
                        class="cursor-pointer hover:text-primary"
                    />
                </a>
            </template>
            <template #form_input_Reason="{ item }">
                <s-input :read-only="!['Submitted'].includes(item.Status)" label="Reason" v-model="item.Reason" class="w-full"></s-input>
            </template>
            <template #form_buttons_2="props">
                <div class="flex flex-row gap-1 ml-[10px]">
                    <s-button v-if="data.showSubmit" label="Submit" class="btn_primary" icon="send" @click="submitLeave" />
                    <s-button v-if="data.showApprove" label="Approve" class="btn_primary" icon="thumb-up" @click="approveReject('Approve')" />
                    <s-button v-if="data.showReject" label="Reject" class="btn_warning" icon="thumb-down" @click="approveReject('Reject')"/>
                </div>
            </template>
            <template #form_footer_2="props">
                <div class="mb-[10px]">
                    <div class="flex flex-row gap-2">
                        <h1>Approvers</h1>
                        <s-button icon="plus" size="8" class="btn_primary" @click="addApprovers" 
                            v-if="props.item.Approvers && props.item.Approvers.length > 0" />
                    </div>
                    <div class="mt-[20px] flex flex-col gap-2">
                        <div v-for="(lineApprovers, index) in props.item.Approvers" class="flex flex-row gap-1 align-middle">
                            <div>{{ index+1 }}</div>
                            <div v-if="index!=props.item.Approvers.length-1" class="w-full">
                                <s-input v-model="lineApprovers.UserIDs"
                                    use-list
                                    lookup-url="/iam/user/find"
                                    lookup-key="_id"
                                    :lookup-labels="['DisplayName']"
                                    multiple
                                    class="w-full"
                                />
                            </div>
                            <div v-else class="flex flex-row gap-1 align-middle">
                                <div v-for="userid in lineApprovers.UserIDs" class="p-[4px] bg-slate-200 rounded-[4px]">{{ userid  }}</div>
                            </div>
                            <div v-if="index!=props.item.Approvers.length-1" class="flex gap-1">
                                <mdicon name="plus" @click="addApprovers(index)" class="cursor-pointer" />
                                <mdicon name="minus" @click="removeApprovers(index)" class="cursor-pointer" />
                            </div>
                        </div>
                    </div>
                </div>

                <data-list grid-mode="grid" form-keep-label no-gap title="Leave Balance" 
                    grid-hide-control grid-hide-detail grid-hide-delete grid-hide-select
                    grid-config="/kara/leavebalance/gridconfig" 
                    :grid-read="`/kara/leavebalance/gets?UserID=${props.item.UserID}`"
                >
                </data-list>
            
                <data-list grid-mode="grid" form-keep-label no-gap title="Last Approved Leave Request" 
                    grid-hide-control grid-hide-detail grid-hide-delete grid-hide-select
                    grid-config="/kara/leave/gridconfig" 
                    grid-sort-field="LeaveFrom"
                    grid-sort-direction="descending"
                    :grid-read="`/kara/leave/gets?UserID=${props.item.UserID}&Status=Approved`"
                >
                </data-list>
            </template>
        </data-list>
    </div>
</template>

<script setup>
import { reactive, ref, onMounted, watch, inject } from "vue";
import { authStore } from "@/stores/auth.js";
import { layoutStore } from "@/stores/layout.js";
import { DataList, SButton, SInput, SCard, util, loadFormConfig } from "suimjs";
import GridHeaderFilter from "@/components/common/GridHeaderFilter.vue";

layoutStore().name = "tenant";

const FEATUREID = 'Leave'
const profile = authStore().getRBAC(FEATUREID)


const axios = inject("axios");
const listControl = ref(null);
const gridHeaderFilter = ref(null);

const data = reactive({
    showSave: true,
    showSubmit: false,
    showApprove: false,
    showReject: false,
    customFilter: null,
});

function newRecord(record) {
    data.showSave = true;
    data.showSubmit = false;
    data.showApprove = false;
    data.showReject = false;
    record.LeaveFrom = new Date();
    record.LeaveTo = new Date();
}

function editRecord(record) {
    recalc(record);
}

function addApprovers(index) {
    let frmRecord = listControl.value.getFormRecord();
    
    if (index==undefined) index=0;
    let newApprovers = [];
    if (frmRecord.Approvers.length > 0) {
        for (let i=0;i<frmRecord.Approvers.length;i++) {
            if (i==index) newApprovers.push({UserIDs:[]});
            newApprovers.push(frmRecord.Approvers[i]);
        }
    } else {
        newApprovers.push({UserIDs:[]});
    }
    frmRecord.Approvers = newApprovers;
    listControl.value.setFormRecord(frmRecord);
}


function removeApprovers(index) {
    let frmRecord = listControl.value.getFormRecord();
    
    if (index==undefined) index=0;
    let newApprovers = [];
    if (frmRecord.Approvers.length > 0) {
        for (let i=0;i<frmRecord.Approvers.length;i++) {
            if (i!=index) newApprovers.push(frmRecord.Approvers[i]);
        }
    } 
    frmRecord.Approvers = newApprovers;
    listControl.value.setFormRecord(frmRecord);
}

function onFormFieldChange(field, v1, v2, record) {
    const frmRecord = listControl.value.getFormRecord();
                
    switch (field) {
        case "LeaveTypeID":
            axios.post("/kara/leave/get-leave-type-approvers",{
                LeaveTypeID: v1,
                UserID: frmRecord.UserID
            }).then(r => {
                if (frmRecord.Approvers==null || frmRecord.Approvers==undefined) frmRecord.Approvers = [];
                if (frmRecord.Approvers.length==0) frmRecord.Approvers.push({UserIDs: r.data})
                    else frmRecord.Approvers[frmRecord.Approvers.length-1] = {UserIDs: r.data}; 
                listControl.value.setFormRecord(frmRecord);
            }, e=>{
                if (frmRecord.Approvers==null || frmRecord.Approvers==undefined) frmRecord.Approvers = [];
                listControl.value.setFormRecord(frmRecord);
            })
    }
}

function recalc(record) {
    if (record.Status==undefined || record.Status=="") {
        data.showSave = true;
        data.showSubmit = true;
        return;
    } else if (record.Status=="Draft") {
        listControl.value.setFormMode("edit");
        data.showSave = true;
        data.showSubmit = true;
    } else {
        listControl.value.setFormMode("view");
        data.showSave = false;
        data.showSubmit = false;
    }

    axios.post('/kara/leave/get-screen-stat',record._id).then(r => {
        const buttons = r.data;
        data.showSave = buttons.filter(d => d=='Submit').length > 0;
        data.showSubmit = buttons.filter(d => d=='Submit').length > 0;
        data.showApprove = buttons.filter(d => d=='Approve').length > 0;
        data.showReject = buttons.filter(d => d=='Reject').length > 0;

        if (data.showApprove || data.showReject)  listControl.value.setFormFieldAttr("Reason","readOnly",false)
            else listControl.value.setFormFieldAttr("Reason","readOnly",true);
    }, e=> util.showError(e));
}

function deleteData(record) {
    axios.post('/kara/leave/delete', record).then(
        (r) => {
            listControl.value.refreshGrid();
        },
        (e) => {
            util.showError(e);
        }
    );
}
function submitLeave() {
    const frmRecord = listControl.value.getFormRecord();
    axios.post('/kara/leave/submit', frmRecord._id).then(r => {
        listControl.value.setFormRecord(r.data);
        recalc(r.data);
        util.showInfo(`Leave request ${frmRecord._id} has been submitted for approval`);
    }, e => util.showError(e));
}

function approveReject(op) {
    const frmRecord = listControl.value.getFormRecord();
    if (op=="Reject" && frmRecord.Reason=="") {
        util.showError("Reason is mandatory for rejection");
        return;
    }
    axios.post('/kara/leave/approve', {ID:frmRecord._id, Op:op, Reason:frmRecord.Reason}).then(r => {
        listControl.value.setFormRecord(r.data);
        recalc(r.data);
        util.showInfo(`Leave request ${frmRecord._id} has been ${op} and continue to next process`);
    }, e => util.showError(e));
}

function initNewItemFilter(item) {
    item.UserID = [];
    item.Title = "";
    item.LeaveFrom = null;
    item.LeaveTo = null;
}

function changeFilter(item, filters) {
    if (item.UserID.length > 0) {
        filters.push({
        Op: "$in",
        Field: "UserID",
        Value: [...item.UserID],
        });
    }
    if (item.Keyword && item.Keyword != "") {
        filters.push({
        Op: "$contains",
        Field: "Name",
        Value: [item.Keyword],
        });
    }
    if (item.LeaveFrom != null) {
        filters.push({
        Op: "$gte",
        Field: "LeaveFrom",
        Value: new Date(item.LeaveFrom),
        });
    }
    if (item.LeaveTo != null) {
        filters.push({
        Op: "$lte",
        Field: "LeaveTo",
        Value: new Date(item.LeaveTo),
        });
    }
}

function refreshGrid() {
    listControl.value.refreshGrid();
}

function resetGridHeaderFilter() {
    gridHeaderFilter.value.reset();
}
</script>