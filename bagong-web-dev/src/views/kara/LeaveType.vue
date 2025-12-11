<template>
    <div>
        <data-list class="card" ref="listControl" grid-mode="grid" form-keep-label title="Leave Type" 
            grid-config="/kara/leave/type/gridconfig" 
            grid-read="/kara/leave/type/gets"
            form-config="/kara/leave/type/formconfig"
            form-read="/kara/leave/type/get"
            form-insert="/kara/leave/type/insert"
            form-update="/kara/leave/type/update"
            :form-tabs-edit="['General','Approvers']"
            :grid-hide-new="!profile.canCreate"
            :grid-hide-edit="!profile.canUpdate"
            :grid-hide-delete="!profile.canDelete"
        >
            <template #form_tab_Approvers="props">
                <data-list grid-mode="grid" form-keep-label no-gap 
                    grid-config="/kara/leaveapprovalsetup/gridconfig" 
                    :grid-read="`/kara/leaveapprovalsetup/gets?LeaveTypeID=${props.item._id}`"
                    form-config="/kara/leaveapprovalsetup/formconfig"
                    form-read="/kara/leaveapprovalsetup/get"
                    form-insert="/kara/leaveapprovalsetup/insert"
                    form-update="/kara/leaveapprovalsetup/update"
                    @form-new-data="newSetup"
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
import { DataList, SCard, util, loadFormConfig } from "suimjs";

layoutStore().name = "tenant";


const FEATUREID = 'LeaveType'
const profile = authStore().getRBAC(FEATUREID)

const axios = inject("axios");
const listControl = ref(null);
const data = reactive({});

function newSetup(record) {
    const leaveType = listControl.value.getFormRecord();
    record.LeaveTypeID = leaveType._id; 
}
</script>